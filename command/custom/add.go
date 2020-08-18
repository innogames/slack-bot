package custom

import (
	"fmt"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/slack-go/slack"
)

func (c *command) Add(match matcher.Result, event slack.MessageEvent) {
	alias := match.GetString("alias")
	command := match.GetString("command")

	list := loadList(event)
	list[alias] = command
	storeList(event, list)

	c.slackClient.Reply(
		event,
		fmt.Sprintf("Added command: `%s`. Just use `%s` in future.", command, alias),
	)
}
