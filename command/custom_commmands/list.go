package custom_commmands

import (
	"fmt"
	"strings"

	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
)

func (c command) list(_ matcher.Result, message msg.Message) {
	list := loadList(message)
	if len(list) == 0 {
		c.SendMessage(message, "No commands define yet. Use `add command 'your alias' 'command to execute'`")
		return
	}

	var responseText strings.Builder
	fmt.Fprintf(&responseText, "You defined %d commands:", len(list))
	for alias, command := range list {
		fmt.Fprintf(&responseText, "\n - %s: `%s`", alias, command)
	}

	c.SendMessage(message, responseText.String())
}
