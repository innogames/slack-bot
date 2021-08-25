package custom

import (
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
)

// GetCommand returns a set of all custom commands
func GetCommand(base bot.BaseCommand) bot.Command {
	return command{base}
}

var category = bot.Category{
	Name:        "Custom Commands",
	Description: "Define your own alias commands",
	HelpURL:     "https://github.com/innogames/slack-bot#custom-command",
}

type command struct {
	bot.BaseCommand
}

func (c command) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.WildcardMatcher(c.handle),
		matcher.NewRegexpMatcher("add command '(?P<alias>.*)'( as)? '(?P<command>.*)'", c.add),
		matcher.NewRegexpMatcher("(delete|remove) command '?(?P<alias>.*?)'?", c.delete),
		matcher.NewTextMatcher("list commands", c.list),
	)
}

func (c command) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "list commands",
			Description: "list all your defined custom commands, which are only available for you",
			Category:    category,
			Examples: []string{
				"list commands",
			},
		},
		{
			Command:     "add command '<alias>' '<command>'",
			Description: "add a custom command/alias which is only available for you",
			Category:    category,
			Examples: []string{
				"`add command 'myCommand' as 'trigger job RestoreWorld 7'` -> then just call `myCommand` later",
				"`add command 'build master' 'trigger job Deploy master ; then trigger job DeployClient master'`",
			},
		},
		{
			Command:     "delete command '<alias>'",
			Description: "define a custom alias",
			Category:    category,
			Examples: []string{
				"delete command 'build master'",
			},
		},
	}
}
