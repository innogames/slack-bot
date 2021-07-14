package custom

import (
	"fmt"

	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
)

func (c command) add(match matcher.Result, message msg.Message) {
	alias := match.GetString("alias")
	command := match.GetString("command")

	list := loadList(message)
	list[alias] = command
	storeList(message, list)

	c.SendMessage(
		message,
		fmt.Sprintf("Added command: `%s`. Just use `%s` in future.", command, alias),
	)
}
