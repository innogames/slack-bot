package pullrequest

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/client"
)

// GetCommands returns a list of all available PR watcher (gitlab, github etc) based on the current config
func GetCommands(slackClient client.SlackClient, cfg config.Config) bot.Commands {
	commands := bot.Commands{}

	commands.AddCommand(
		newGitlabCommand(slackClient, cfg),
		newGithubCommand(slackClient, cfg),
		newBitbucketCommand(slackClient, cfg),
	)

	return commands
}
