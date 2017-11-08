package command

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/config"
	"github.com/innogames/slack-bot/mocks"
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBotLog(t *testing.T) {
	testFile := "test.log"

	slackClient := mocks.SlackClient{}
	cfg := config.Config{}
	cfg.Logger.File = testFile
	cfg.AdminUsers = []string{
		"UADMIN",
	}

	botLog := NewBotLogCommand(&slackClient, cfg)

	command := bot.Commands{}
	command.AddCommand(botLog)

	t.Run("invalid command", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "log log log"

		actual := command.Run(event)
		assert.Equal(t, false, actual)
	})

	t.Run("display log without history", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "bot log"
		event.User = "UADMIN"

		slackClient.On("Reply", event, "No logs so far")

		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})
}
