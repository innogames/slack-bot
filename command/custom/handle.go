package custom

import (
	"fmt"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/nlopes/slack"
	"strings"
)

type handleCommand struct {
	baseCommand
}

func (c *handleCommand) GetMatcher() matcher.Matcher {
	return matcher.NewConditionalMatcher(c.Execute)
}

func (c *handleCommand) Execute(event slack.MessageEvent) bool {
	var commands string

	list := loadList(event)
	if commands = list[event.Text]; commands == "" {
		return false
	}

	c.slackClient.Reply(event, fmt.Sprintf("executing command: `%s`", commands))
	for _, command := range strings.Split(commands, ";") {
		newMessage := event
		newMessage.Text = command
		client.InternalMessages <- newMessage
	}

	return true
}
