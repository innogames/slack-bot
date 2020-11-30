package variables

import (
	"fmt"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/bot/msg"
)

func (c *command) Delete(match matcher.Result, message msg.Message) {
	name := match.GetString("name")

	list := loadList(message.GetUser())
	delete(list, name)
	storeList(message, list)

	c.slackClient.SendMessage(message, fmt.Sprintf("Okay, I deleted variable: `%s`", name))
}
