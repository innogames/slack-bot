package variables

import (
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
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
			Command: "set variable",
			Examples: []string{
				"set variable 'server' 'foo.prod.local'",
			},
			Category: category,
		},
		{
			Command: "delete variable",
			Examples: []string{
				"remove variable 'server'",
			},
			Category: category,
		},
		{
			Command:  "list variables",
			Examples: []string{"list variables"},
			Category: category,
		},
	}
}

var category = bot.Category{
	Name:        "Custom Variables",
	Description: "usable e.g. in 'commands' or 'crons'",
	HelpURL:     "https://github.com/innogames/slack-bot#custom-variables",
}
