package pullrequest

import (
	gojira "github.com/andygrunwald/go-jira"
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/command/jira"
	log "github.com/sirupsen/logrus"
)

var category = bot.Category{
	Name:        "Pull Request",
	Description: "Track the state of pull/merge requests",
	HelpURL:     "https://github.com/innogames/slack-bot#pull-requests",
}

// GetCommands returns a list of all available PR watcher (gitlab, github, bitbucket etc) based on the current config
func GetCommands(base bot.BaseCommand, cfg *config.Config) bot.Commands {
	commands := bot.Commands{}

	// optional Jira client to resolve the priority/severity of the ticket referenced by a PR
	var jiraClient *gojira.Client
	if cfg.Jira.IsEnabled() {
		if client, err := jira.GetClient(&cfg.Jira); err == nil {
			jiraClient = client
		} else {
			// non-fatal: PR watching still works, just without severity reactions
			log.Warnf("error while initializing Jira client for PR severity: %s", err)
		}
	}

	commands.AddCommand(
		newGitlabCommand(base, cfg, jiraClient),
		newGithubCommand(base, cfg, jiraClient),
		newBitbucketCommand(base, cfg, jiraClient),
	)

	return commands
}
