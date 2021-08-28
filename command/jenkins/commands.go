package jenkins

import (
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/client/jenkins"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// help category to group all Jenkins command
var category = bot.Category{
	Name:        "Jenkins",
	Description: "Interact with Jenkins jobs: Trigger builds, check job status or watch builds",
	HelpURL:     "https://github.com/innogames/slack-bot#jenkins-optional",
}

// base command to access Slack+Jenkins directly
type jenkinsCommand struct {
	bot.BaseCommand
	jenkins jenkins.Client
}

// GetCommands will return a list of available Jenkins commands...if the config is set!
func GetCommands(cfg config.Jenkins, base bot.BaseCommand) bot.Commands {
	var commands bot.Commands

	if !cfg.IsEnabled() {
		return commands
	}

	jenkinsClient, err := jenkins.GetClient(cfg)
	if err != nil {
		log.Error(errors.Wrap(err, "Error while getting Jenkins client"))
		return commands
	}

	jenkinsBase := jenkinsCommand{
		base,
		jenkinsClient,
	}

	commands.AddCommand(
		newTriggerCommand(jenkinsBase, cfg.Jobs),
		newJobWatcherCommand(jenkinsBase),
		newBuildWatcherCommand(jenkinsBase),
		newStatusCommand(jenkinsBase, cfg.Jobs),
		newNodesCommand(jenkinsBase, cfg),
		newRetryCommand(jenkinsBase, cfg.Jobs),
		newIdleWatcherCommand(jenkinsBase),
	)

	return commands
}
