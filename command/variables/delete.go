package variables

import (
	"fmt"

	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
)

func (c *command) delete(match matcher.Result, message msg.Message) {
	name := match.GetString("name")

	list := loadList(message.GetUser())
	delete(list, name)
	storeList(message, list)

	c.SendMessage(message, fmt.Sprintf("Okay, I deleted variable: `%s`", name))
}
