package openai

import (
	"fmt"
	"strings"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/storage"
	"github.com/innogames/slack-bot/v2/bot/util"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

// only use the last X messages as context for further requests
const (
	historySize = 10
	storageKey  = "chatgpt"
)

// GetCommands if enable, register the openai commands
func GetCommands(base bot.BaseCommand, config *config.Config) bot.Commands {
	var commands bot.Commands

	cfg := loadConfig(config)
	if !cfg.IsEnabled() {
		return commands
	}

	commands.AddCommand(
		&chatGPTCommand{
			base,
			cfg,
		},
	)

	return commands
}

type chatGPTCommand struct {
	bot.BaseCommand
	cfg Config
}

func (c *chatGPTCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewPrefixMatcher("openai", c.startConversation),
		matcher.NewPrefixMatcher("chatgpt", c.startConversation),
		matcher.WildcardMatcher(c.reply),
	)
}

// bot function which is called, when the user started a new conversation with openat/chatgpt
func (c *chatGPTCommand) startConversation(match matcher.Result, message msg.Message) {
	text := match.GetString(util.FullMatch)

	// Call the API with a fresh history
	storageIdentifier := getIdentifier(message.GetChannel(), message.GetTimestamp())
	messageHistory := make([]ChatMessage, 0)
	if c.cfg.InitialSystemMessage != "" {
		messageHistory = append(messageHistory, ChatMessage{
			Role:    roleSystem,
			Content: c.cfg.InitialSystemMessage,
		})
	}

	// add a 🗒️️ emoji to the thread message to make clear that it's using chatgpt magic
	c.AddReaction(":spiral_note_pad:", message)

	c.callAndStore(messageHistory, storageIdentifier, message, text)
}

// bot function which is called, when the user replied in a openai/chatgpt thread
func (c *chatGPTCommand) reply(message msg.Ref, text string) bool {
	if message.GetThread() == "" {
		// We're only interested in thread replies
		return false
	}

	// Load the chat history from storage.
	identifier := getIdentifier(message.GetChannel(), message.GetThread())

	var messages []ChatMessage
	storage.Read(storageKey, identifier, &messages)
	if len(messages) == 0 {
		return false
	}

	// Call the API and send the last messages as history to give a proper context
	c.callAndStore(messages, identifier, message, text)

	return true
}

// call the GPT-3 API, sends the response to the user, and stores the updated chat history.
func (c *chatGPTCommand) callAndStore(messages []ChatMessage, storageIdentifier string, message msg.Ref, text string) {
	// Append the new user input to the message list.
	messages = append(messages, ChatMessage{
		Role:    roleUser,
		Content: text,
	})

	// Use a custom coffee emoji reaction while we wait for OpenAI to respond...could take some time
	c.AddReaction(":coffee:", message)
	defer c.RemoveReaction(":coffee:", message)

	resp, err := CallChatGPT(c.cfg, messages)
	if err != nil {
		c.ReplyError(message, fmt.Errorf("openai error: %w", err))
		return
	}

	// Send the response to the user.
	c.SendMessage(
		message,
		resp.GetMessage().Content,
		slack.MsgOptionTS(message.GetTimestamp()),
	)

	// Store the last X chat history entries for further questions
	messages = append(messages, resp.GetMessage())
	if len(messages) > historySize {
		messages = messages[len(messages)-historySize:]
	}
	err = storage.Write(storageKey, storageIdentifier, messages)
	if err != nil {
		log.Warnf("Error while storing openai history: %s", err)
	}
}

// create a unique storage key which is stable for all messages in a thread
func getIdentifier(channel string, threadTS string) string {
	identifier := fmt.Sprintf("%s-%s", channel, threadTS)

	return strings.ReplaceAll(identifier, ".", "_")
}

func (c *chatGPTCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "openai *question*",
			Description: "Starts a chatgpt/openai conversation in a new thread",
			Examples: []string{
				"openai whats 1+1?",
				"chatgpt whats 1+1?",
			},
		},
	}
}
