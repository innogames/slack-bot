package variables

import (
	"fmt"

	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/slack-go/slack"
)

func (c *command) List(match matcher.Result, event slack.MessageEvent) {
	list := loadList(event.User)
	if len(list) == 0 {
		c.slackClient.Reply(event, "No variables define yet. Use `add variable 'defaultServer' 'beta'`")
		return
	}

	responseText := fmt.Sprintf("You defined %d variables:", len(list))
	for name, value := range list {
		responseText += fmt.Sprintf("\n - %s: `%s`", name, value)
	}

	c.slackClient.Reply(event, responseText)
}
