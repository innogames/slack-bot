package command

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/slack-go/slack"
)

// NewSendMessageCommand is able to send a message to any user/channel
func NewSendMessageCommand(slackClient client.SlackClient) bot.Command {
	return &sendMessageCommand{slackClient}
}

type sendMessageCommand struct {
	slackClient client.SlackClient
}

func (c *sendMessageCommand) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher("send message( to)? (?P<fullChannel><(?P<type>[#@])(?P<receiver>\\w+)(?i:|[^>]*)?>) (?P<text>.*)", c.SendMessage)
}

func (c *sendMessageCommand) SendMessage(match matcher.Result, event slack.MessageEvent) {
	text := match.GetString("text")
	if event.User != "" {
		text = fmt.Sprintf("Text from <@%s>: %s", event.User, text)
	}

	if match.GetString("type") == "#" {
		// send to channel
		newEvent := slack.MessageEvent{}
		newEvent.Channel = match.GetString("receiver")
		c.slackClient.Reply(newEvent, text)
	} else {
		c.slackClient.SendToUser(match.GetString("receiver"), text)
	}

	c.slackClient.Reply(
		event,
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
