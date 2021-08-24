package command

import (
	"math/rand"
	"strings"
	"text/template"
	"time"

	"github.com/innogames/slack-bot/v2/bot/util"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
)

// NewRandomCommand will reply a random entry
func NewRandomCommand(base bot.BaseCommand) bot.Command {
	random := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec

	return &randomCommand{base, random}
}

type randomCommand struct {
	bot.BaseCommand
	random *rand.Rand
}

func (c *randomCommand) GetMatcher() matcher.Matcher {
	return matcher.NewPrefixMatcher("random", c.getRandom)
}

func (c *randomCommand) getRandom(match matcher.Result, message msg.Message) {
	optionsString := match.GetString(util.FullMatch)
	if optionsString == "" {
		c.SendMessage(message, "You have to pass more arguments")
		return
	}
	options := strings.Split(optionsString, " ")

	randomOption := c.pickRandom(options)

	c.SendMessage(message, randomOption)
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
			Command:     "random <whatever...>",
			Description: "returns a random element of the given parameters",
			Examples: []string{
				"random 1 2 3 3 4",
				"random peter paul tom",
				"random pizza pasta kebab",
			},
		},
	}
}
