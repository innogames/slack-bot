package custom

import (
	"fmt"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/nlopes/slack"
)

type addCommand struct {
	baseCommand
}

func (c *addCommand) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher("add command '(?P<alias>.*)'( as)? '(?P<command>.*)'", c.Run)
}

func (c *addCommand) Run(match matcher.Result, event slack.MessageEvent) {
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
