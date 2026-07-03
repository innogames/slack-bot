// Package example is a minimal reference plugin, showing how to extend the slack-bot
// with own commands: implement bot.Command(s), register them via bot.RegisterPlugin in
// an init() function and activate the plugin with a blank import in your main package.
// It is intentionally NOT part of plugins/all - see docs/plugins.md.
package example

import (
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/innogames/slack-bot/v2/client"
)

// Config of the example plugin, read from the "plugins: example:" config section
type Config struct {
	Prefix string `mapstructure:"prefix"`
}

type echoCommand struct {
	bot.BaseCommand
	cfg Config
}

func (c *echoCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewPrefixMatcher(c.cfg.Prefix, c.reply),
	)
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
		Name: "example",
		Init: getCommands,
	})
}

func getCommands(slackClient client.SlackClient, cfg config.Config) bot.Commands {
	commands := bot.Commands{}

	pluginCfg := Config{Prefix: "echo"}
	if err := cfg.LoadPlugin("example", &pluginCfg); err != nil {
		return commands
	}

	commands.AddCommand(&echoCommand{
		BaseCommand: bot.BaseCommand{SlackClient: slackClient},
		cfg:         pluginCfg,
	})

	return commands
}
