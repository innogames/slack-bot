package command

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/client"
	"math/rand"
	"strings"
	"text/template"
	"time"
)

// NewRandomCommand will reply a random entry
func NewRandomCommand(slackClient client.SlackClient) bot.Command {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))

	return &randomCommand{slackClient, random}
}

type randomCommand struct {
	slackClient client.SlackClient
	random      *rand.Rand
}

func (c *randomCommand) GetMatcher() matcher.Matcher {
	return matcher.NewPrefixMatcher("random", c.GetRandom)
}

func (c *randomCommand) GetRandom(match matcher.Result, message msg.Message) {
	optionsString := match.MatchedString()
	if optionsString == "" {
		c.slackClient.SendMessage(message, "You have to pass more arguments")
		return
	}
	options := strings.Split(optionsString, " ")

	randomOption := c.pickRandom(options)

	c.slackClient.SendMessage(message, randomOption)
}

func (c *randomCommand) pickRandom(list []string) string {
	return list[c.random.Intn(len(list))]
}

func (c *randomCommand) GetTemplateFunction() template.FuncMap {
	return template.FuncMap{
		"random": c.pickRandom,
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
