package variables

import (
	"fmt"
	"github.com/innogames/slack-bot/bot/msg"

	"github.com/innogames/slack-bot/bot/matcher"
)

func (c *command) Add(match matcher.Result, message msg.Message) {
	name := match.GetString("name")
	value := match.GetString("value")

	list := loadList(message.GetUser())
	list[name] = value
	storeList(message, list)

	c.slackClient.SendMessage(
		message,
		fmt.Sprintf("Added variable: `%s` = `%s`.", name, value),
	)
}
