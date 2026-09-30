// Package hello is an example plugin which replies with a configurable greeting
package hello

import (
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
)

// Config is defined in "plugins.hello.config"
type Config struct {
	Greeting string `mapstructure:"greeting"`
}

type helloCommand struct {
	bot.BaseCommand
	cfg Config
}

func (c *helloCommand) GetMatcher() matcher.Matcher {
	return matcher.NewTextMatcher("hello", func(_ matcher.Result, message msg.Message) {
		c.SendMessage(message, c.cfg.Greeting)
	})
}

func (c *helloCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "hello",
			Description: "replies with a greeting",
		},
	}
}

func init() {
	bot.RegisterPlugin(bot.Plugin{
		Name: "hello",
		Setup: func(ctx *bot.PluginContext) (bot.Commands, error) {
			commands := bot.Commands{}

			cfg := Config{Greeting: "Hello!"}
			if err := ctx.LoadConfig(&cfg); err != nil {
				return commands, err
			}

			commands.AddCommand(&helloCommand{ctx.BaseCommand(), cfg})

			return commands, nil
		},
	})
}
