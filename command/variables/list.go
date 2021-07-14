package variables

import (
	"fmt"

	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
)

func (c *command) list(match matcher.Result, message msg.Message) {
	list := loadList(message.GetUser())
	if len(list) == 0 {
		c.SendMessage(message, "No variables define yet. Use `add variable 'defaultServer' 'beta'`")
		return
	}

	responseText := fmt.Sprintf("You defined %d variables:", len(list))
	for name, value := range list {
		responseText += fmt.Sprintf("\n - %s: `%s`", name, value)
	}

	c.SendMessage(message, responseText)
}
