package custom_commmands

import (
	"fmt"
	"strings"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"gopkg.in/yaml.v3"
)

func (c command) export(match matcher.Result, message msg.Message) {
	list := loadList(message)
	if len(list) == 0 {
		c.SendMessage(message, "No commands define yet. Use `add command 'your alias' 'command to execute'`")
		return
	}

	commands := make([]config.Command, 0, len(list))
	for alias, command := range list {
		commands = append(commands, config.Command{
			Trigger:  alias,
			Commands: strings.Split(command, ";"),
		})
	}

	out, _ := yaml.Marshal(commands)
	c.SendMessage(message, fmt.Sprintf("```%s```", out))
}
