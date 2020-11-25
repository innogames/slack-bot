package jenkins

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/client/jenkins"
	"github.com/sirupsen/logrus"
)

var category = bot.Category{
	Name:        "Jenkins",
	Description: "Interact with Jenkins jobs: Trigger builds, check job status or watch builds",
	HelpUrl:     "https://github.com/innogames/slack-bot#jenkins",
}

// GetCommands will return a list of available Jenkins commands...if the config is set!
func GetCommands(cfg config.Jenkins, slackClient client.SlackClient, logger *logrus.Logger) bot.Commands {
	var commands bot.Commands

	if !cfg.IsEnabled() {
		return commands
	}

	jenkinsClient, err := jenkins.GetClient(cfg)
	if err != nil {
		logger.Error(err)
		return commands
	}
	commands.AddCommand(
		newTriggerCommand(jenkinsClient, slackClient, cfg.Jobs, logger),
		newJobWatcherCommand(jenkinsClient, slackClient, logger),
		newBuildWatcherCommand(jenkinsClient, slackClient),
		newStatusCommand(jenkinsClient, slackClient, cfg.Jobs),
		newNodesCommand(jenkinsClient, slackClient),
		newRetryCommand(jenkinsClient, slackClient, cfg.Jobs, logger),
	)

	return commands
}
