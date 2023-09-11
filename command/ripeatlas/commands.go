package ripeatlas

import (
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
)

var category = bot.Category{
	Name:        "RIPE Atlas",
	Description: "Run queries against the RIPE Atlas API to debug network issues",
}

func GetCommands(base bot.BaseCommand, config *config.Config) bot.Commands {
	var commands bot.Commands

	cfg := loadConfig(config)
	if !cfg.IsEnabled() {
		return commands
	}

	commands.AddCommand(
		&creditsCommand{base, cfg},
		&tracerouteCommand{base, cfg},
	)

	return commands
}
