package custom

import (
	"fmt"

	"github.com/innogames/slack-bot/bot/matcher"
	msg "github.com/innogames/slack-bot/bot/msg"
)

func (c command) Delete(match matcher.Result, message msg.Message) {
	alias := match.GetString("alias")

	list := loadList(message)
	delete(list, alias)
	storeList(message, list)

	c.SendMessage(message, fmt.Sprintf("Okay, I deleted command: `%s`", alias))
}
