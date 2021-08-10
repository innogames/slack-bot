package admin

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
)

func TestPing(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	botLog := newPingCommand(base)

	command := bot.Commands{}
	command.AddCommand(botLog)

	t.Run("test ping", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "ping"

		slackClient.On("SendMessage", message, "PONG").Return("")

		actual := command.Run(message)
		assert.True(t, actual)
	})
}
