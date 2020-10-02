package variables

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
)

// GetCommand returns a set of all commands to manage user specific variables
func GetCommand(slackClient client.SlackClient) bot.Command {
	return &command{slackClient}
}

type command struct {
	slackClient client.SlackClient
}

func (c *command) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewRegexpMatcher("(add|set) variable '?(?P<name>.*?)'? '?(?P<value>.*?)'?", c.Add),
		matcher.NewRegexpMatcher("(delete|remove) variable '?(?P<name>.*?)'?", c.Delete),
		matcher.NewTextMatcher("list variables", c.List),
	)
}

func (c *command) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:  "custom variables",
			Examples: []string{},
		},
	}
}
