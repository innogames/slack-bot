package custom

import (
	"fmt"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/client"
	"strings"
)

func (c command) Handle(ref msg.Ref, text string) bool {
	var commands string

	list := loadList(ref)
	if commands = list[text]; commands == "" {
		return false
	}

	c.SendMessage(ref, fmt.Sprintf("executing command: `%s`", commands))
	for _, command := range strings.Split(commands, ";") {
		client.InternalMessages <- ref.WithText(command)
	}

	return true
}
