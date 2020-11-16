package pullrequest

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/client"
	"github.com/sirupsen/logrus"
)

// GetCommands returns a list of all available PR watcher (gitlab, github, bitbucket etc) based on the current config
func GetCommands(slackClient client.SlackClient, cfg config.Config, logger *logrus.Logger) bot.Commands {
	commands := bot.Commands{}

	commands.AddCommand(
		newGitlabCommand(slackClient, cfg, logger),
		newGithubCommand(slackClient, cfg, logger),
		newBitbucketCommand(slackClient, cfg, logger),
	)

	return commands
}
