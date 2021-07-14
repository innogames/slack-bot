package jenkins

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
)

// just a test helper to setup all needed mocks etc
func getTestJenkinsCommand() (*mocks.SlackClient, *mocks.Client, jenkinsCommand) {
	slackClient := &mocks.SlackClient{}
	jenkinsClient := &mocks.Client{}

	base := jenkinsCommand{bot.BaseCommand{SlackClient: slackClient}, jenkinsClient}

	return slackClient, jenkinsClient, base
}

func TestGetCommands(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	t.Run("Jenkins is not active", func(t *testing.T) {
		cfg := config.Jenkins{}
		commands := GetCommands(cfg, base)
		assert.Equal(t, 0, commands.Count())
	})

	t.Run("Jenkins is active", func(t *testing.T) {
		cfg := config.Jenkins{}
		cfg.Host = "http://ci.jenkins-ci.org"
		commands := GetCommands(cfg, base)
		assert.Equal(t, 7, commands.Count())
	})
}
