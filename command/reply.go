package command

import (
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/slack-go/slack"
)

// NewReplyCommand is a command to reply a message in current thread/channel
func NewReplyCommand(base bot.BaseCommand) bot.Command {
	return &replyCommand{base}
}

type replyCommand struct {
	bot.BaseCommand
}

func (c *replyCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewPrefixMatcher("reply", c.reply),
		matcher.NewPrefixMatcher("comment", c.commentInNewThread),
	)
}

func (c *replyCommand) reply(match matcher.Result, message msg.Message) {
	text := match.GetString(util.FullMatch)
	if text == "" {
		return
	}

	c.SendMessage(message, text)
}

// comment in (new) thread
func (c *replyCommand) commentInNewThread(match matcher.Result, message msg.Message) {
	text := match.GetString(util.FullMatch)
	if text == "" {
		return
	}

	c.SendMessage(message, text, slack.MsgOptionTS(message.GetTimestamp()))
}

func (c *replyCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "reply <text>",
			Description: "just reply the given message",
			Category:    helperCategory,
			Examples: []string{
				"reply Hello, how are you?",
			},
		},
		{
			Command:     "comment <text>",
			Description: "comment the message in a new thread on this message",
			Category:    helperCategory,
			Examples: []string{
				"comment Hello, how are you?",
			},
		},
	}
}
