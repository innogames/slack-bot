package variables

import (
	"fmt"

	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/slack-go/slack"
)

func (c *command) Add(match matcher.Result, event slack.MessageEvent) {
	name := match.GetString("name")
	value := match.GetString("value")

	list := loadList(event.User)
	list[name] = value
	storeList(event, list)

	c.slackClient.Reply(
		event,
		fmt.Sprintf("Added variable: `%s` = `%s`.", name, value),
	)
}
