package gitlab

import (
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	log "github.com/sirupsen/logrus"
	"gitlab.com/gitlab-org/api/client-go"
)

var category = bot.Category{
	Name:        "GitLab",
	Description: "Interact with GitLab pipelines and jobs",
}

// GetCommands returns GitLab-related commands, only if GitLab is configured
func GetCommands(base bot.BaseCommand, cfg *config.Config) bot.Commands {
	var commands bot.Commands

	if cfg.Gitlab.AccessToken == "" || cfg.Gitlab.Host == "" {
		return commands
	}

	options := gitlab.WithBaseURL(cfg.Gitlab.Host)
	gitlabClient, err := gitlab.NewClient(cfg.Gitlab.AccessToken, options)
	if err != nil {
		log.Errorf("Error creating GitLab client: %s", err)
		return commands
	}

	baseCmd := gitlabCommand{
		BaseCommand: base,
		api:         &realGitlabAPI{client: gitlabClient},
		host:        cfg.Gitlab.Host,
	}

	commands.AddCommand(
		newNotifyCommand(baseCmd),
	)

	return commands
}
