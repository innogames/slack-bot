package admin

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/mocks"
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
		message := msg.Message{}
		message.Text = "log log log"

		actual := command.Run(message)
		assert.Equal(t, false, actual)
	})

	t.Run("display log without history", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "bot log"
		message.User = "UADMIN"

		slackClient.On("SendMessage", message, "No logs so far").Return("")

		actual := command.Run(message)
		assert.Equal(t, true, actual)
	})
}
