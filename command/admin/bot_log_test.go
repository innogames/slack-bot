package admin

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
)

func TestBotLog(t *testing.T) {
	testFile := "test.log"

	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	cfg := &config.Config{}
	cfg.Logger.File = testFile
	cfg.AdminUsers = []string{
		"UADMIN",
	}

	botLog := NewBotLogCommand(base, cfg)

	command := bot.Commands{}
	command.AddCommand(botLog)

	t.Run("invalid command", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "log log log"

		actual := command.Run(message)
		assert.False(t, actual)
	})

	t.Run("display log without history", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "bot log"
		message.User = "UADMIN"

		slackClient.On("SendMessage", message, "No logs so far").Return("")

		actual := command.Run(message)
		assert.True(t, actual)
	})
}
