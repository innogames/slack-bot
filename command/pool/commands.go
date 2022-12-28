package pool

import (
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/client"
)

// GetCommands will return a list of available Pool commands...if the config is set!
func GetCommands(cfg *config.Pool, slackClient client.SlackClient) bot.Commands {
	var commands bot.Commands

	if !cfg.IsEnabled() {
		return commands
	}

	p := GetNewPool(cfg)

	commands.AddCommand(
		newPoolCommands(slackClient, cfg, p),
	)

	return commands
}

var category = bot.Category{
	Name:        "Pool",
	Description: "Lock/Unlock/Manage Resources of a Pool",
	HelpURL:     "https://github.com/innogames/slack-bot",
}
