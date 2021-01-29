package command

import (
	"testing"

	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/storage"
	"github.com/innogames/slack-bot/mocks"
	"github.com/stretchr/testify/assert"
)

func TestAddButton(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	slackClient.On("CanHandleInteractions").Return(true)

	base := bot.BaseCommand{SlackClient: slackClient}

	command := bot.Commands{}
	command.AddCommand(NewAddButtonCommand(base))

	t.Run("add link", func(t *testing.T) {
		message := msg.Message{}
		message.Text = `add button "test" "reply it works"`

		expected := `[{"type":"actions","elements":[{"type":"button","text":{"type":"plain_text","text":"test","emoji":true},"action_id":"id","value":"token-0"}]}]`

		mocks.AssertSlackBlocks(t, slackClient, message, expected)

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
