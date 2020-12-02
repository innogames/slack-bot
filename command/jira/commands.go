package jira

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/client"
	log "github.com/sirupsen/logrus"
)

// GetCommands will return a list of available Jira commands...if the config is set!
func GetCommands(cfg *config.Jira, slackClient client.SlackClient) bot.Commands {
	var commands bot.Commands

	if !cfg.IsEnabled() {
		return commands
	}

	jira, err := client.GetJiraClient(cfg)
	if err != nil {
		log.Error(err)
		return commands
	}

	commands.AddCommand(
		newJiraCommand(jira, slackClient, cfg),
		newWatchCommand(jira, slackClient, cfg),
		newCommentCommand(jira, slackClient, cfg),
	)

	return commands
}

var category = bot.Category{
	Name:        "Jira",
	Description: "Search/Watch Jira tickets",
	HelpURL:     "https://github.com/innogames/slack-bot#jira",
}
