package command

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/mocks"
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

		expected := `[{"type":"actions","elements":[{"type":"button","text":{"type":"plain_text","text":"test","emoji":true},"action_id":"id","value":"reply it works"}]}]`

		mocks.AssertSlackBlocks(t, slackClient, message, expected)

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Test help", func(t *testing.T) {
		help := command.GetHelp()
		assert.Equal(t, 1, len(help))
	})
}
