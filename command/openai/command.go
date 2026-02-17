package openai

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/stats"
	"github.com/innogames/slack-bot/v2/bot/storage"
	"github.com/innogames/slack-bot/v2/bot/util"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

const (
	storageKey = "chatgpt"
)

var linkRe = regexp.MustCompile(`<?https://(.*?)\.slack\.com/archives/(?P<channel>\w+)/p(?P<timestamp>\d{16})>?`)

// GetCommands if enable, register the openai commands
func GetCommands(base bot.BaseCommand, config *config.Config) bot.Commands {
	var commands bot.Commands

	cfg := loadConfig(config)
	if !cfg.IsEnabled() {
		return commands
	}

	commands.AddCommand(
		&openaiCommand{
			base,
			cfg,
		},
	)

	return commands
}

type openaiCommand struct {
	bot.BaseCommand
	cfg Config
}

func (c *openaiCommand) GetMatcher() matcher.Matcher {
	matchers := []matcher.Matcher{
		matcher.NewPrefixMatcher("openai", c.newConversation),
		matcher.NewPrefixMatcher("chatgpt", c.newConversation),
		matcher.NewPrefixMatcher("dalle", c.dalleGenerateImage),
		matcher.NewPrefixMatcher("dall-e", c.dalleGenerateImage),
		matcher.NewPrefixMatcher("generate image", c.dalleGenerateImage),
		matcher.WildcardMatcher(c.reply),
	}

	// if configured evaluate the given command text as openai request
	if c.cfg.UseAsFallback {
		matchers = append(matchers, matcher.WildcardMatcher(c.startConversation))
	}

	return matcher.NewGroupMatcher(matchers...)
}

// bot function which is called, when the user started a new conversation with openai/chatgpt
func (c *openaiCommand) newConversation(match matcher.Result, message msg.Message) {
	text := match.GetString(util.FullMatch)
	c.startConversation(message.MessageRef, text)
}

func (c *openaiCommand) startConversation(message msg.Ref, text string) bool {
	// Parse hashtags and get cleaned text
	cleanText, hashtagOptions := ParseHashtags(text)

	messageHistory := make([]ChatMessage, 0)

	if c.cfg.InitialSystemMessage != "" {
		messageHistory = append(messageHistory, ChatMessage{
			Role:    roleSystem,
			Content: c.cfg.InitialSystemMessage,
		})
	}

	// Handle #message-history hashtag
	if hashtagOptions.MessageHistory > 0 {
		channelMessages, err := c.getChannelHistory(message.GetChannel(), hashtagOptions.MessageHistory)
		if err != nil {
			c.ReplyError(message, fmt.Errorf("can't load channel history: %w", err))
			return true
		}

		// Add channel history as context
		messageHistory = append(messageHistory, ChatMessage{
			Role:    roleSystem,
			Content: fmt.Sprintf("Recent channel conversation history (last %d messages, including threads):", hashtagOptions.MessageHistory),
		})

		for _, channelMsg := range channelMessages {
			var prefix string
			// Check if this is a thread reply by comparing ThreadTimestamp with Timestamp
			if channelMsg.ThreadTimestamp != "" && channelMsg.ThreadTimestamp != channelMsg.Timestamp {
				prefix = "  ↳ " // Indent thread replies
			}
			messageHistory = append(messageHistory, ChatMessage{
				Role:    roleUser,
				Content: fmt.Sprintf("%sUser <@%s> wrote: %s", prefix, channelMsg.User, channelMsg.Text),
			})
		}
	}

	var storageIdentifier string
	if message.GetThread() != "" {
		// "openai" was triggered within a existing thread. -> fetch the whole thread history as context
		threadMessages, err := c.GetThreadMessages(message)
		if err != nil {
			c.ReplyError(message, fmt.Errorf("can't load thread messages: %w", err))
			return true
		}

		messageHistory = append(messageHistory, ChatMessage{
			Role:    roleSystem,
			Content: "This is a Slack bot receiving a slack thread s context, using slack user ids as identifiers. Please use user mentions in the format <@U123456>",
		})

		for _, threadMessage := range threadMessages {
			content := threadMessage.Text
			if len(threadMessage.Files) > 0 {
				content += c.loadTextAttachments(threadMessage.Files)
			}
			messageHistory = append(messageHistory, ChatMessage{
				Role:    roleUser,
				Content: fmt.Sprintf("User <@%s> wrote: %s", threadMessage.User, content),
			})
		}
		storageIdentifier = getIdentifier(message.GetChannel(), message.GetThread())
		if c.cfg.LogTexts {
			log.Infof("openai thread context: %s", messageHistory)
		}
	} else if linkRe.MatchString(cleanText) {
		// a link to another thread was posted -> use this messages as context
		link := linkRe.FindStringSubmatch(cleanText)
		cleanText = linkRe.ReplaceAllString(cleanText, "")

		relatedMessage := msg.MessageRef{
			Channel: link[2],
			Thread:  link[3][0:10] + "." + link[3][10:],
		}
		threadMessages, err := c.GetThreadMessages(relatedMessage)
		if err != nil {
			c.ReplyError(message, fmt.Errorf("can't load thread messages: %w", err))
			return true
		}

		for _, threadMessage := range threadMessages {
			content := threadMessage.Text
			if len(threadMessage.Files) > 0 {
				content += c.loadTextAttachments(threadMessage.Files)
			}
			messageHistory = append(messageHistory, ChatMessage{
				Role:    roleUser,
				Content: fmt.Sprintf("User <@%s> wrote: %s", threadMessage.User, content),
			})
		}

		storageIdentifier = getIdentifier(message.GetChannel(), message.GetTimestamp())
	} else {
		// start a new thread with a fresh history
		storageIdentifier = getIdentifier(message.GetChannel(), message.GetTimestamp())
	}

	c.callAndStore(messageHistory, storageIdentifier, message, cleanText, hashtagOptions)
	return true
}

// bot function which is called when the user replied in a openai/chatgpt thread
func (c *openaiCommand) reply(message msg.Ref, text string) bool {
	if message.GetThread() == "" {
		// We're only interested in thread replies
		return false
	}

	// Parse hashtags from reply message
	cleanText, hashtagOptions := ParseHashtags(text)

	// Load the chat history from storage.
	identifier := getIdentifier(message.GetChannel(), message.GetThread())

	var messages []ChatMessage
	err := storage.Read(storageKey, identifier, &messages)
	if err != nil || len(messages) == 0 {
		// no "openai thread"
		return false
	}

	// Call the API and send the last messages as history to give a proper context
	c.callAndStore(messages, identifier, message, cleanText, hashtagOptions)

	return true
}

// call the GPT-3 API, sends the response to the user, and stores the updated chat history.
func (c *openaiCommand) callAndStore(messages []ChatMessage, storageIdentifier string, message msg.Ref, inputText string, options HashtagOptions) {
	// Append the actual user input to the message list.
	messages = append(messages, ChatMessage{
		Role:    roleUser,
		Content: inputText,
	})

	// Create a custom config based on hashtag options
	customCfg := c.cfg

	// Apply model override if specified
	if options.Model != "" {
		customCfg.Model = options.Model
	}

	// Apply reasoning effort if specified
	if options.ReasoningEffort == "none" {
		// User explicitly requested no reasoning effort with #no-thinking
		customCfg.ReasoningEffort = ""
	} else if options.ReasoningEffort != "" {
		customCfg.ReasoningEffort = options.ReasoningEffort
	}

	messages, inputTokens, truncatedMessages := truncateMessages(customCfg.Model, messages)
	if truncatedMessages > 0 {
		// Build message options based on NoThread setting
		msgOptions := []slack.MsgOption{}
		if !options.NoThread || message.GetThread() != "" {
			// Normal behavior: reply in thread
			msgOptions = append(msgOptions, slack.MsgOptionTS(message.GetTimestamp()))
		}
		c.SendMessage(
			message,
			fmt.Sprintf("Note: The token length of %d exceeded! %d messages were not sent", getMaxTokensForModel(customCfg.Model), truncatedMessages),
			msgOptions...,
		)
	}

	// wait for the full event stream in the background to not block other user requests
	go func() {
		// Use a bulb emoji reaction while we wait for OpenAI to start responding (thinking/reasoning phase)
		c.AddReaction(":bulb:", message)

		var speechBalloonAdded bool
		defer func() {
			if speechBalloonAdded {
				c.RemoveReaction(":speech_balloon:", message)
			} else {
				// If we never got a response, remove the bulb
				c.RemoveReaction(":bulb:", message)
			}
		}()

		startTime := time.Now()

		// Determine if streaming should be used based on hashtag option
		useStreaming := !options.NoStreaming

		// Use customCfg instead of c.cfg
		response, err := CallChatGPT(customCfg, messages, useStreaming)
		if err != nil {
			c.ReplyError(message, fmt.Errorf("openai error: %w", err))
			return
		}

		// Create a dummy message which gets updated every X seconds
		// Build message options based on NoThread setting
		msgOptions := []slack.MsgOption{}
		if !options.NoThread || message.GetThread() != "" {
			// Normal behavior: reply in thread
			msgOptions = append(msgOptions, slack.MsgOptionTS(message.GetTimestamp()))
		}
		replyRef := c.SendMessage(
			message,
			":bulb: thinking...",
			msgOptions...,
		)

		// Initialize message chunker with 3500 char limit per chunk
		// Pass NoThread option to chunker so it knows whether to use thread timestamps
		useNoThread := options.NoThread && message.GetThread() == ""
		chunker := newMessageChunker(c.BaseCommand, message, replyRef, 3500, useNoThread)

		var lastUpdate time.Time
		var dirty bool
		var outputTokens int
		for delta := range response {
			// On first token received (including empty deltas), switch from bulb (thinking) to speech_balloon (streaming response)
			if !speechBalloonAdded {
				c.RemoveReaction(":bulb:", message)
				c.AddReaction(":speech_balloon:", message)
				speechBalloonAdded = true
			}

			// Only append non-empty content to avoid issues with role-only deltas
			if delta != "" {
				chunker.appendContent(delta)
				outputTokens++
				dirty = true
				if chunker.getTotalLength() > 0 && lastUpdate.Add(c.cfg.UpdateInterval).Before(time.Now()) {
					lastUpdate = time.Now()
					chunker.updateMessages()
					dirty = false
				}
			}
		}

		// update with the final message to make sure everything is formatted properly
		if dirty {
			chunker.updateMessages()
		}

		// Store the last X chat history entries for further questions
		messages = append(messages, ChatMessage{
			Role:    roleAssistant,
			Content: chunker.getFullText(),
		})

		// truncate the history if needed to the last X messages
		if len(messages) > c.cfg.HistorySize {
			messages = messages[len(messages)-c.cfg.HistorySize:]
		}

		err = storage.Write(storageKey, storageIdentifier, messages)
		if err != nil {
			log.Warnf("Error while storing openai history: %s", err)
		}

		// log some stats in the end
		stats.IncreaseOne("openai_calls")
		stats.Increase("openai_input_tokens", inputTokens)
		stats.Increase("openai_output_tokens", outputTokens)

		logFields := log.Fields{
			"input_tokens":  inputTokens,
			"output_tokens": outputTokens,
			"model":         customCfg.Model,
		}
		if options.ReasoningEffort != "" {
			logFields["reasoning_effort"] = customCfg.ReasoningEffort
		}
		if options.MessageHistory > 0 {
			logFields["message_history_count"] = options.MessageHistory
		}
		if c.cfg.LogTexts {
			logFields["input_text"] = inputText
			logFields["output_text"] = chunker.getFullText()
		}

		log.WithFields(logFields).Infof(
			"Openai call took %s with %d context messages.",
			util.FormatDuration(time.Since(startTime)),
			len(messages),
		)

		// Send debug message if debug option is enabled
		if options.Debug {
			debugMessage := fmt.Sprintf(
				"🔍 *Debug Info:*\n"+
					"• Model: `%s`\n"+
					"• Input tokens: `%d`\n"+
					"• Output tokens: `%d`\n"+
					"• Total time: `%s`\n"+
					"• Reasoning effort: `%s`",
				customCfg.Model,
				inputTokens,
				outputTokens,
				util.FormatDuration(time.Since(startTime)),
				customCfg.ReasoningEffort,
			)

			// Build message options based on NoThread setting
			debugMsgOptions := []slack.MsgOption{}
			if !options.NoThread || message.GetThread() != "" {
				// Normal behavior: reply in thread
				debugMsgOptions = append(debugMsgOptions, slack.MsgOptionTS(message.GetTimestamp()))
			}

			c.SendMessage(
				message,
				debugMessage,
				debugMsgOptions...,
			)
		}
	}()
}

// create a unique storage key which is stable for all messages in a thread
func getIdentifier(channel string, threadTS string) string {
	identifier := fmt.Sprintf("%s-%s", channel, threadTS)

	return strings.ReplaceAll(identifier, ".", "_")
}

// GetTemplateFunction makes "chatgpt" available as template function for custom commands
func (c *openaiCommand) GetTemplateFunction() template.FuncMap {
	return template.FuncMap{
		"openai": func(input string) string {
			message := []ChatMessage{
				{
					Role:    roleUser,
					Content: input,
				},
			}
			responses, err := CallChatGPT(c.cfg, message, false)
			if err != nil {
				return err.Error()
			}
			finalMessage := <-responses

			return strings.Trim(finalMessage, "\n")
		},
	}
}

// loadTextAttachments downloads text file attachments and returns them formatted
func (c *openaiCommand) loadTextAttachments(files []slack.File) string {
	var result strings.Builder
	for _, file := range files {
		if !strings.HasPrefix(file.Mimetype, "text/") {
			log.Infof("Skipping attachment %s: mimetype is %s", file.Name, file.Mimetype)
			continue
		}

		var buf bytes.Buffer
		log.Infof("Downloading attachment %s", file.Name)

		if err := c.GetFile(file.URLPrivate, &buf); err != nil {
			log.Warnf("Failed to download attachment %s: %v", file.Name, err)
			continue
		}

		fmt.Fprintf(&result, " <Attachment filename=\"%s\">%s</Attachment>", file.Name, buf.String())
	}
	return result.String()
}

// getChannelHistory fetches the last N messages from a channel (including thread messages and text attachments)
func (c *openaiCommand) getChannelHistory(channel string, count int) ([]slack.Message, error) {
	params := &slack.GetConversationHistoryParameters{
		ChannelID: channel,
		Limit:     count,
	}

	response, err := c.GetConversationHistory(params)
	if err != nil {
		return nil, err
	}

	// Collect main channel messages and their thread replies
	allMessages := make([]slack.Message, 0)
	for _, message := range response.Messages {
		if len(message.Files) > 0 {
			message.Text += c.loadTextAttachments(message.Files)
		}
		allMessages = append(allMessages, message)

		// If this message has replies, fetch the thread messages
		if message.ReplyCount > 0 {
			threadRef := msg.MessageRef{
				Channel:   channel,
				Timestamp: message.Timestamp,
				Thread:    message.Timestamp,
			}
			threadMessages, err := c.GetThreadMessages(threadRef)
			if err != nil {
				log.Warnf("Failed to fetch thread messages for %s: %v", message.Timestamp, err)
				continue
			}

			// Skip the first message as it's the parent (already added)
			if len(threadMessages) > 1 {
				for i := 1; i < len(threadMessages); i++ {
					if len(threadMessages[i].Files) > 0 {
						threadMessages[i].Text += c.loadTextAttachments(threadMessages[i].Files)
					}
				}
				allMessages = append(allMessages, threadMessages[1:]...)
			}
		}
	}

	// Reverse to get chronological order (API returns newest first)
	for i := range len(allMessages) / 2 {
		j := len(allMessages) - i - 1
		allMessages[i], allMessages[j] = allMessages[j], allMessages[i]
	}

	return allMessages, nil
}

func (c *openaiCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "openai <question>",
			Description: "Starts a chatgpt/openai conversation in a new thread. Text file attachments are automatically included in the context. Supports hashtags for advanced options: \n- #model-<name> (e.g., #model-gpt-4o), \n- #high-thinking/#medium-thinking/#low-thinking/#no-thinking for reasoning control\n- #message-history or #message-history-<N> to include recent channel messages as context\n- #no-streaming to disable streaming and get the full response at once\n- #no-thread to reply directly without creating a thread\n- #debug to show debug information about the request",
			Category:    category,
			Examples: []string{
				"openai why is the sky blue?",
				"openai #high-thinking #model-gpt-5-pro explain quantum computing",
				"openai #message-history-20 what did we discuss about the deployment?",
				"chatgpt #model-gpt-5-nano #low-thinking #message-history quick summary please",
				"openai #no-streaming give me a detailed explanation",
				"openai #no-thread quick question without a thread",
				"openai #debug analyze this code for performance issues",
			},
		},
		{
			Command:     "dalle <prompt>",
			Description: "Generates an image with Dall-E",
			Category:    category,
			Examples: []string{
				"dalle high resolution image of a sunset, painted by a robot",
				"dall-e high resolution image of a sunset, painted by a robot",
			},
		},
	}
}

// help category to group all AI command
var category = bot.Category{
	Name:        "AI",
	Description: "AI support for commands, using openai (GPT or DALL-E)",
	HelpURL:     "https://github.com/innogames/slack-bot#openaichatgptdall-e-integration",
}
