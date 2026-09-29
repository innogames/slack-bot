// Package example is a minimal reference plugin, showing how to extend the slack-bot with own commands:
// implement bot.Command(s) and register them via bot.RegisterPlugin in an init() function. See docs/plugins.md
package example

import (
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/util"
)

// Config of the example plugin, defined in "plugins.example.config"
type Config struct {
	Prefix string `mapstructure:"prefix"`
}

type echoCommand struct {
	bot.BaseCommand
	cfg Config
}

func (c *echoCommand) GetMatcher() matcher.Matcher {
	return matcher.NewPrefixMatcher(c.cfg.Prefix, c.reply)
}

func (c *echoCommand) reply(match matcher.Result, message msg.Message) {
	text := match.GetString(util.FullMatch)
	if text == "" {
		return
	}

	c.SendMessage(message, text)
}

func (c *echoCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     c.cfg.Prefix + " <text>",
			Description: "echoes the given text",
			Examples:    []string{c.cfg.Prefix + " hello world"},
		},
	}
}

func init() {
	bot.RegisterPlugin(bot.Plugin{
		Name:  "example",
		Setup: setup,
	})
}

// setup is called when the plugin is enabled in the "plugins" config section
func setup(ctx *bot.PluginContext) (bot.Commands, error) {
	commands := bot.Commands{}

	cfg := Config{Prefix: "echo"}
	if err := ctx.LoadConfig(&cfg); err != nil {
		return commands, err
	}

	commands.AddCommand(&echoCommand{
		BaseCommand: ctx.BaseCommand(),
		cfg:         cfg,
	})

	return commands, nil
}
