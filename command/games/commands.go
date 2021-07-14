package games

import (
	"github.com/innogames/slack-bot/v2/bot"
)

var category = bot.Category{
	Name:        "Games",
	Description: "Just for fun...",
}

// GetCommands will return a list of available games commands
func GetCommands(base bot.BaseCommand) bot.Commands {
	var commands bot.Commands

	commands.AddCommand(
		NewNumberGuesserCommand(base),
		NewQuizCommand(base),
	)

	return commands
}
