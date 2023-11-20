package openai

import (
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
	messageHistory := make([]ChatMessage, 0)

	if c.cfg.InitialSystemMessage != "" {
		messageHistory = append(messageHistory, ChatMessage{
			Role:    roleSystem,
			Content: c.cfg.InitialSystemMessage,
		})
	}

	var storageIdentifier string
	if message.GetThread() != "" {
		// "openai" was triggerd within a existing thread. -> fetch the whole thread history as context
		threadMessages, err := c.SlackClient.GetThreadMessages(message)
		if err != nil {
			c.ReplyError(message, fmt.Errorf("can't load thread messages: %w", err))
			return true
		}

		messageHistory = append(messageHistory, ChatMessage{
			Role:    roleSystem,
			Content: "This is a Slack bot receiving a slack thread s context, using slack user ids as identifiers. Please use user mentions in the format <@U123456>",
		})

		for _, threadMessage := range threadMessages {
			messageHistory = append(messageHistory, ChatMessage{
				Role:    roleUser,
				Content: fmt.Sprintf("User <@%s> wrote: %s", threadMessage.User, threadMessage.Text),
			})
		}
		storageIdentifier = getIdentifier(message.GetChannel(), message.GetThread())
		if c.cfg.LogTexts {
			log.Infof("openai thread context: %s", messageHistory)
		}
	} else if linkRe.MatchString(text) {
		// a link to another thread was posted -> use this messages as context
		link := linkRe.FindStringSubmatch(text)
		text = linkRe.ReplaceAllString(text, "")

		relatedMessage := msg.MessageRef{
			Channel: link[2],
			Thread:  link[3][0:10] + "." + link[3][10:],
		}
		threadMessages, err := c.SlackClient.GetThreadMessages(relatedMessage)
		if err != nil {
			c.ReplyError(message, fmt.Errorf("can't load thread messages: %w", err))
			return true
		}

		for _, threadMessage := range threadMessages {
			messageHistory = append(messageHistory, ChatMessage{
				Role:    roleUser,
				Content: fmt.Sprintf("User <@%s> wrote: %s", threadMessage.User, threadMessage.Text),
			})
		}

		storageIdentifier = getIdentifier(message.GetChannel(), message.GetTimestamp())
	} else {
		// start a new thread with a fresh history
		storageIdentifier = getIdentifier(message.GetChannel(), message.GetTimestamp())
	}

	c.callAndStore(messageHistory, storageIdentifier, message, text)
	return true
}

// bot function which is called when the user replied in a openai/chatgpt thread
func (c *openaiCommand) reply(message msg.Ref, text string) bool {
	if message.GetThread() == "" {
		// We're only interested in thread replies
		return false
	}

	// Load the chat history from storage.
	identifier := getIdentifier(message.GetChannel(), message.GetThread())

	var messages []ChatMessage
	err := storage.Read(storageKey, identifier, &messages)
	if err != nil || len(messages) == 0 {
		// no "openai thread"
		return false
	}

	// Call the API and send the last messages as history to give a proper context
	c.callAndStore(messages, identifier, message, text)

	return true
}

// call the GPT-3 API, sends the response to the user, and stores the updated chat history.
func (c *openaiCommand) callAndStore(messages []ChatMessage, storageIdentifier string, message msg.Ref, inputText string) {
	// Append the actual user input to the message list.
	messages = append(messages, ChatMessage{
		Role:    roleUser,
		Content: inputText,
	})

	messages, inputTokens, truncatedMessages := truncateMessages(c.cfg.Model, messages)
	if truncatedMessages > 0 {
		c.SendMessage(
			message,
			fmt.Sprintf("Note: The token length of %d exceeded! %d messages were not sent", getMaxTokensForModel(c.cfg.Model), truncatedMessages),
			slack.MsgOptionTS(message.GetTimestamp()),
		)
	}

	// wait for the full event stream in the background to not block other user requests
	go func() {
		// Use a custom coffee emoji reaction while we wait for the full OpenAI response
		c.AddReaction(":coffee:", message)
		defer c.RemoveReaction(":coffee:", message)

		startTime := time.Now()

		response, err := CallChatGPT(c.cfg, messages, true)
		if err != nil {
			c.ReplyError(message, fmt.Errorf("openai error: %w", err))
			return
		}

		// Create a dummy message which gets updated every X seconds
		replyRef := c.SendMessage(
			message,
			"...",
			slack.MsgOptionTS(message.GetTimestamp()),
		)

		responseText := strings.Builder{}
		var lastUpdate time.Time
		var dirty bool
		for delta := range response {
			responseText.WriteString(delta)
			dirty = true
			if responseText.Len() > 0 && lastUpdate.Add(c.cfg.UpdateInterval).Before(time.Now()) {
				lastUpdate = time.Now()

				c.SendMessage(
					message,
					responseText.String(),
					slack.MsgOptionTS(message.GetTimestamp()),
					slack.MsgOptionUpdate(replyRef),
				)
				dirty = false
			}
		}

		// update with the final message to make sure everything is formatted properly
		if dirty {
			c.SendMessage(
				message,
				responseText.String(),
				slack.MsgOptionTS(message.GetTimestamp()),
				slack.MsgOptionUpdate(replyRef),
			)
		}

		// Store the last X chat history entries for further questions
		messages = append(messages, ChatMessage{
			Role:    roleAssistant,
			Content: responseText.String(),
		})
		if len(messages) > c.cfg.HistorySize {
			messages = messages[len(messages)-c.cfg.HistorySize:]
		}
		err = storage.Write(storageKey, storageIdentifier, messages)
		if err != nil {
			log.Warnf("Error while storing openai history: %s", err)
		}

		// log some stats in the end
		outputTokens := estimateTokensForMessage(responseText.String())
		stats.IncreaseOne("openai_calls")
		stats.Increase("openai_input_tokens", inputTokens)
		stats.Increase("openai_output_tokens", outputTokens)

		logFields := log.Fields{
			"input_tokens":  inputTokens,
			"output_tokens": outputTokens,
			"model":         c.cfg.Model,
		}
		if c.cfg.LogTexts {
			logFields["input_text"] = inputText
			logFields["output_text"] = responseText.String()
		}

		log.WithFields(logFields).Infof(
			"Openai call took %s with %d context messages.",
			util.FormatDuration(time.Since(startTime)),
			len(messages),
		)
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

func (c *openaiCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "openai <question>",
			Description: "Starts a chatgpt/openai conversation in a new thread",
			Category:    category,
			Examples: []string{
				"openai whats 1+1?",
				"chatgpt whats 1+1?",
			},
		},
		{
			Command:     "dalle <prompt>",
			Description: "Generates an image with Dall-E",
			Category:    category,
			Examples: []string{
				"dalle high resolution image of a sunset, painted by a robot",
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
