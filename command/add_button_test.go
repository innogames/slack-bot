package command

import (
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/storage"
	"github.com/stretchr/testify/mock"
	"testing"

	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/mocks"
	"github.com/stretchr/testify/assert"
)

func TestAddButton(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	cfg := config.Server{}
	cfg.Listen = "0.0.0.0:1234"
	cfg.SigningSecret = "iamsecret"

	command := bot.Commands{}
	command.AddCommand(NewAddButtonCommand(base, cfg))

	t.Run("add link", func(t *testing.T) {
		message := msg.Message{}
		message.Text = `add button "test" "reply it works"`

		slackClient.On("SendMessage", message, "", mock.Anything).Return("")

		actual := command.Run(message)
		assert.True(t, actual)

		storedKeys, err := storage.GetKeys("interactions")
		assert.Nil(t, err)
		assert.Len(t, storedKeys, 1)
	})

	t.Run("Test help", func(t *testing.T) {
		help := command.GetHelp()
		assert.Equal(t, 1, len(help))
	})
}
