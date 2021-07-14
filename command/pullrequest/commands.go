package pullrequest

import (
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
)

var category = bot.Category{
	Name:        "Pull Request",
	Description: "Track the state of pull/merge requests",
	HelpURL:     "https://github.com/innogames/slack-bot#pull-requests",
}

// GetCommands returns a list of all available PR watcher (gitlab, github, bitbucket etc) based on the current config
func GetCommands(base bot.BaseCommand, cfg *config.Config) bot.Commands {
	commands := bot.Commands{}

	commands.AddCommand(
		newGitlabCommand(base, cfg),
		newGithubCommand(base, cfg),
		newBitbucketCommand(base, cfg),
	)

	return commands
}
