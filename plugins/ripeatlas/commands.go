package ripeatlas

import (
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/client"
)

var category = bot.Category{
	Name:        "RIPE Atlas",
	Description: "Run queries against the RIPE Atlas API to debug network issues",
}

func init() {
	bot.RegisterPlugin(bot.Plugin{
		Name: "ripeatlas",
		Init: getCommands,
	})
}

func getCommands(slackClient client.SlackClient, config config.Config) bot.Commands {
	var commands bot.Commands

	cfg := loadConfig(&config)
	if !cfg.IsEnabled() {
		return commands
	}

	base := bot.BaseCommand{SlackClient: slackClient}
	commands.AddCommand(
		&creditsCommand{base, cfg},
		&tracerouteCommand{base, cfg},
	)

	return commands
}
