package variables

import (
	"fmt"
	"github.com/innogames/slack-bot/bot/msg"

	"github.com/innogames/slack-bot/bot/matcher"
)

func (c *command) List(match matcher.Result, message msg.Message) {
	list := loadList(message.GetUser())
	if len(list) == 0 {
		c.slackClient.SendMessage(message, "No variables define yet. Use `add variable 'defaultServer' 'beta'`")
		return
	}

	responseText := fmt.Sprintf("You defined %d variables:", len(list))
	for name, value := range list {
		responseText += fmt.Sprintf("\n - %s: `%s`", name, value)
	}

	c.slackClient.SendMessage(message, responseText)
}
