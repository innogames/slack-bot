package games

import (
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
)

var category = bot.Category{
	Name:        "Games",
	Description: "Just for fun...",
}

// GetCommands will return a list of available games commands
func GetCommands(base bot.BaseCommand, cfg *config.Config) bot.Commands {
	var commands bot.Commands

	gameConfig := loadConfig(cfg)
	if !gameConfig.Enabled {
		return commands
	}

	commands.AddCommand(
		NewNumberGuesserCommand(base),
		NewQuizCommand(base),
	)

	return commands
}
