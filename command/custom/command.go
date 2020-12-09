package custom

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
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
		matcher.WildcardMatcher(c.Handle),
		matcher.NewRegexpMatcher("add command '(?P<alias>.*)'( as)? '(?P<command>.*)'", c.Add),
		matcher.NewRegexpMatcher("(delete|remove) command '?(?P<alias>.*?)'?", c.Delete),
		matcher.NewTextMatcher("list commands", c.List),
	)
}

// todo separate commands + separate help entries!
func (c command) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "custom commands",
			Description: "Define command aliases which are just available for you. You can use a `;` to separate single commands",
			Category:    category,
			Examples: []string{
				"`list commands`",
				"`add command 'myCommand' as 'trigger job RestoreWorld 7'` -> then just call `myCommand` later",
				"`add command 'build master' 'trigger job Deploy master ; then trigger job DeployClient master'`",
				"`delete command 'build master'`",
			},
		},
	}
}
