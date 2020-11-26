package games

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/client"
)

var category = bot.Category{
	Name:        "Games",
	Description: "Just fun...",
}

// GetCommands will return a list of available games commands
func GetCommands(slackClient client.SlackClient) bot.Commands {
	var commands bot.Commands

	commands.AddCommand(
		NewNumberGuesserCommand(slackClient),
		NewQuizCommand(slackClient),
	)

	return commands
}
