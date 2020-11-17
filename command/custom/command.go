package custom

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
)

// GetCommand returns a set of all custom commands
func GetCommand(slackClient client.SlackClient) bot.Command {
	return command{slackClient}
}

type command struct {
	slackClient client.SlackClient
}

func (c command) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewConditionalMatcher(c.Handle),
		matcher.NewRegexpMatcher("add command '(?P<alias>.*)'( as)? '(?P<command>.*)'", c.Add),
		matcher.NewRegexpMatcher("(delete|remove) command '?(?P<alias>.*?)'?", c.Delete),
		matcher.NewTextMatcher("list commands", c.List),
	)
}

func (c command) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "custom commands",
			Description: "Define command aliases which are just available for you. You can use a `;` to separate single commands",
			Examples: []string{
				"`list commands`",
				"`add command 'myCommand' as 'trigger job RestoreWorld 7'` -> then just call `myCommand` later",
				"`add command 'build master' 'trigger job Deploy master ; then trigger job DeployClient master'`",
				"`delete command 'build master'`",
			},
		},
	}
}
