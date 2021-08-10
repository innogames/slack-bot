package admin

import (
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
)

// GetCommands will return a list of available admin commands
func GetCommands(base bot.BaseCommand, cfg *config.Config) bot.Commands {
	var commands bot.Commands

	commands.AddCommand(
		newStatsCommand(base, cfg),
		newBotLogCommand(base, cfg),
		newPingCommand(base),
	)

	return commands
}
