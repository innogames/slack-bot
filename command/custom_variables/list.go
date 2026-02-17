package custom_variables

import (
	"fmt"
	"strings"

	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
)

func (c *command) list(_ matcher.Result, message msg.Message) {
	list := loadList(message.GetUser())
	if len(list) == 0 {
		c.SendMessage(message, "No variables define yet. Use `add variable 'defaultServer' 'beta'`")
		return
	}

	var responseText strings.Builder
	fmt.Fprintf(&responseText, "You defined %d variables:", len(list))
	for name, value := range list {
		fmt.Fprintf(&responseText, "\n - %s: `%s`", name, value)
	}

	c.SendMessage(message, responseText.String())
}
