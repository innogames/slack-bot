package command

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/slack-go/slack"
	"math/rand"
	"strings"
	"text/template"
)

// NewRandomCommand will reply a random entry
func NewRandomCommand(slackClient client.SlackClient) bot.Command {
	return &randomCommand{slackClient}
}

type randomCommand struct {
	slackClient client.SlackClient
}

func (c *randomCommand) GetMatcher() matcher.Matcher {
	return matcher.NewPrefixMatcher("random", c.GetRandom)
}

func (c *randomCommand) GetRandom(match matcher.Result, event slack.MessageEvent) {
	optionsString := match.MatchedString()
	if optionsString == "" {
		c.slackClient.Reply(event, "You have to pass more arguments")
		return
	}
	options := strings.Split(optionsString, " ")

	randomOption := pickRandom(options)

	c.slackClient.Reply(event, randomOption)
}

func pickRandom(list []string) string {
	return list[rand.Intn(len(list))]
}

func (c *randomCommand) GetTemplateFunction() template.FuncMap {
	return template.FuncMap{
		"random": pickRandom,
	}
}

func (c *randomCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "random",
			Description: "returns a random element of the given parameters",
			Examples: []string{
				"random 1 2 3 3 4",
				"random peter paul tom",
				"random pizza pasta kebab",
			},
		},
	}
}
