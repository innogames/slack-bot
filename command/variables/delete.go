package variables

import (
	"fmt"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/slack-go/slack"
)

func (c *command) Delete(match matcher.Result, event slack.MessageEvent) {
	name := match.GetString("name")

	list := loadList(event.User)
	delete(list, name)
	storeList(event, list)

	c.slackClient.Reply(event, fmt.Sprintf("Okay, I deleted variable: `%s`", name))
}
