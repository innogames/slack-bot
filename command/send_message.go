package command

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/client"
)

// NewSendMessageCommand is able to send a message to any user/channel
func NewSendMessageCommand(slackClient client.SlackClient) bot.Command {
	return &sendMessageCommand{slackClient}
}

type sendMessageCommand struct {
	slackClient client.SlackClient
}

func (c *sendMessageCommand) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher(`send message( to)? (?P<fullChannel><(?P<type>[#@])(?P<receiver>\w+)(?i:|[^>]*)?>) (?P<text>.*)`, c.SendMessage)
}

func (c *sendMessageCommand) SendMessage(match matcher.Result, message msg.Message) {
	text := match.GetString("text")
	if message.GetUser() != "" {
		text = fmt.Sprintf("Text from <@%s>: %s", message.User, text)
	}

	if match.GetString("type") == "#" {
		// send to channel
		newEvent := msg.Message{}
		newEvent.Channel = match.GetString("receiver")
		c.slackClient.SendMessage(newEvent, text)
	} else {
		c.slackClient.SendToUser(match.GetString("receiver"), text)
	}

	c.slackClient.SendMessage(
		message,
		fmt.Sprintf("I'll send `%s` to %s", match.GetString("text"), match.GetString("fullChannel")),
	)
}

func (c *sendMessageCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "send message",
			Description: "sends a message to given user/channel",
			Examples: []string{
				"send message #dev-backend PANIC MODE!!!",
				"send message to @username please take a look in #general",
			},
		},
	}
}
