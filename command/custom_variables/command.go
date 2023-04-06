package custom_variables

import (
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
)

// GetCommand returns a set of all commands to manage user specific variables
func GetCommand(base bot.BaseCommand, cfg *config.Config) bot.Command {
	return command{base, cfg}
}

type command struct {
	bot.BaseCommand
	cfg *config.Config
}

func (c command) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewRegexpMatcher("(add|set) variable '?(?P<name>.*?)'? '?(?P<value>.*?)'?", c.add),
		matcher.NewRegexpMatcher("(delete|remove) variable '?(?P<name>.*?)'?", c.delete),
		matcher.NewTextMatcher("list variables", c.list),
	)
}

func (c command) IsEnabled() bool {
	cfg := loadConfig(c.cfg)

	return cfg.Enabled
}

func (c command) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "set variable <key> <value>",
			Description: "Set a custom variable for just the current user",
			Examples: []string{
				"set variable 'server' 'foo.prod.local'",
			},
			Category: category,
		},
		{
			Command:     "delete variable <key>",
			Description: "Remove a custom variable (for current user only)",
			Examples: []string{
				"remove variable 'server'",
			},
			Category: category,
		},
		{
			Command:     "list variables",
			Description: "List you custom variables",
			Examples:    []string{"list variables"},
			Category:    category,
		},
	}
}

var category = bot.Category{
	Name:        "Custom Variables",
	Description: "usable e.g. in 'commands' or 'crons'",
	HelpURL:     "https://github.com/innogames/slack-bot#custom-variables",
}
