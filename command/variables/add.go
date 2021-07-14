package variables

import (
	"fmt"

	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
)

func (c *command) add(match matcher.Result, message msg.Message) {
	name := match.GetString("name")
	value := match.GetString("value")

	list := loadList(message.GetUser())
	list[name] = value
	storeList(message, list)

	c.SendMessage(
		message,
		fmt.Sprintf("Added variable: `%s` = `%s`.", name, value),
	)
}
