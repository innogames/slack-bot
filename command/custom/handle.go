package custom

import (
	"fmt"
	"strings"

	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/client"
)

func (c command) Handle(ref msg.Ref, text string) bool {
	var commands string

	list := loadList(ref)
	if commands = list[text]; commands == "" {
		return false
	}

	c.SendMessage(ref, fmt.Sprintf("executing command: `%s`", commands))
	for _, command := range strings.Split(commands, ";") {
		client.HandleMessage(ref.WithText(command))
	}

	return true
}
