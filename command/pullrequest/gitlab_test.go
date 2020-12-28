package pullrequest

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGitlab(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	cfg := config.Config{}
	cfg.Gitlab.AccessToken = "https://gitlab.example.com"
	cfg.Gitlab.AccessToken = "0815"

	commands := bot.Commands{}
	cmd := newGitlabCommand(base, cfg).(command)
	commands.AddCommand(cmd)

	t.Run("help", func(t *testing.T) {
		help := commands.GetHelp()
		assert.Equal(t, 1, len(help))
	})

}
