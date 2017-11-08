package custom

import (
	"fmt"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/nlopes/slack"
)

type deleteCommand struct {
	baseCommand
}

func (c *deleteCommand) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher("(delete|remove) command '?(?P<alias>.*?)'?", c.Run)
}

func (c *deleteCommand) Run(match matcher.Result, event slack.MessageEvent) {
	alias := match.GetString("alias")

	list := loadList(event)
	delete(list, alias)
	storeList(event, list)

	c.slackClient.Reply(event, fmt.Sprintf("Okay, I deleted command: `%s`", alias))
}
