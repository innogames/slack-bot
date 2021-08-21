package main

import (
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/client"
)

type exampleCommand struct {
	client.SlackClient
}

func (c exampleCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewTextMatcher("test", func(match matcher.Result, message msg.Message) {
			c.SendMessage(message, "yep, it works!")
		}),
	)
}
