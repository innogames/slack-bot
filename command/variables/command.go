package variables

import (
	"github.com/innogames/slack-bot.v2/bot"
	"github.com/innogames/slack-bot.v2/bot/matcher"
)

// GetCommand returns a set of all commands to manage user specific variables
func GetCommand(base bot.BaseCommand) bot.Command {
	return command{base}
}

type command struct {
	bot.BaseCommand
}

func (c command) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewRegexpMatcher("(add|set) variable '?(?P<name>.*?)'? '?(?P<value>.*?)'?", c.add),
		matcher.NewRegexpMatcher("(delete|remove) variable '?(?P<name>.*?)'?", c.delete),
		matcher.NewTextMatcher("list variables", c.list),
	)
}

func (c command) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:  "custom variables",
			Examples: []string{},
		},
	}
}
