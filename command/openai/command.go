package openai

import (
	"fmt"
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/storage"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/slack-go/slack"
)

const storageKey = "chatgpt"

func GetCommands(base bot.BaseCommand, cfg *config.Config) bot.Commands {
	var commands bot.Commands
	var openaiCfg OpenAIConfig

	cfg.LoadCustom("openai", &openaiCfg)
	if !openaiCfg.IsEnabled() {
		return commands
	}

	commands.AddCommand(
		newChatGPTCommand(base, openaiCfg),
	)

	fmt.Println("Init openai")

	return commands
}

// newPingCommand just prints a PING with the needed time from client->slack->bot server
func newChatGPTCommand(base bot.BaseCommand, cfg OpenAIConfig) bot.Command {
	return &chatGPTCommand{
		base,
		cfg,
	}
}

type chatGPTCommand struct {
	bot.BaseCommand
	cfg OpenAIConfig
}

func (c *chatGPTCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewPrefixMatcher("openai", c.startConversation), // todo how to start a conversation?
		matcher.WildcardMatcher(c.allReplies),
	)
}

func (c *chatGPTCommand) startConversation(match matcher.Result, message msg.Message) {
	c.AddReaction("☕", message)
	defer c.RemoveReaction("☕", message)

	messages := []ChatMessage{
		{
			Role:    roleUser,
			Content: match.GetString(util.FullMatch),
		},
	}

	resp, err := CallChatGPT(c.cfg, messages)
	if err != nil {
		c.ReplyError(message, fmt.Errorf("openai error: %w", err))
		return
	}

	c.SendMessage(
		message,
		resp.GetRecentMessage(),
		slack.MsgOptionTS(message.GetTimestamp()),
	)

	// store the history
	messages = append(messages, resp.Choices[0].Message)
	storage.Write(storageKey, message.GetUniqueKey(), messages)
}

func (c *chatGPTCommand) allReplies(message msg.Ref, text string) bool {
	if message.GetThread() == "" {
		// We're only interested in thread replies
		return false
	}

	// load history
	var messages []ChatMessage
	storage.Read(storageKey, message.GetUniqueKey(), &messages) // todo which key?
	if len(messages) == 0 {
		return false
	}

	// todo copy&paste
	messages = append(messages, ChatMessage{
		Role:    roleUser,
		Content: text,
	})

	c.AddReaction("☕", message)
	defer c.RemoveReaction("☕", message)

	resp, err := CallChatGPT(c.cfg, messages)
	if err != nil {
		c.ReplyError(message, fmt.Errorf("openai error: %w", err))
		return true
	}

	// store the new history
	messages = append(messages, resp.Choices[0].Message)

	storage.Write(storageKey, message.GetUniqueKey(), messages)

	return true
}

func (c *chatGPTCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "openai: *question*",
			Description: "Starts a chatgpt/openai conversation in a new thread",
			Examples: []string{
				"openai: whats 1+1?",
			},
		},
	}
}
