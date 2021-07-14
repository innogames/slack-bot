package custom

import (
	"fmt"

	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
)

func (c command) delete(match matcher.Result, message msg.Message) {
	alias := match.GetString("alias")

	list := loadList(message)
	delete(list, alias)
	storeList(message, list)

	c.SendMessage(message, fmt.Sprintf("Okay, I deleted command: `%s`", alias))
}
