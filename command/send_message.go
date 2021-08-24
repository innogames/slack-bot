package command

import (
	"fmt"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
)

// NewSendMessageCommand is able to send a message to any user/channel
func NewSendMessageCommand(base bot.BaseCommand) bot.Command {
	return &sendMessageCommand{base}
}

type sendMessageCommand struct {
	bot.BaseCommand
}

func (c *sendMessageCommand) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher(`send message( to)? (?P<fullChannel><(?P<type>[#@])(?P<receiver>\w+)(?i:|[^>]*)?>) (?P<text>.*)`, c.sendMessage)
}

func (c *sendMessageCommand) sendMessage(match matcher.Result, message msg.Message) {
	text := match.GetString("text")
	if message.GetUser() != "" {
		text = fmt.Sprintf("Text from <@%s>: %s", message.User, text)
	}

	if match.GetString("type") == "#" {
		// send to channel
		newEvent := msg.Message{}
		newEvent.Channel = match.GetString("receiver")
		c.SlackClient.SendMessage(newEvent, text)
	} else {
		c.SendToUser(match.GetString("receiver"), text)
	}

	c.SlackClient.SendMessage(
		message,
		fmt.Sprintf("I'll send `%s` to %s", match.GetString("text"), match.GetString("fullChannel")),
	)
}

func (c *sendMessageCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "send message <message> to <who>",
			Description: "sends a message to given user/channel",
			Examples: []string{
				"send message #dev-backend PANIC MODE!!!",
				"send message to @username please take a look in #general",
			},
		},
	}
}
