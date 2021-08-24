package command

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
)

func TestAddLink(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	command := bot.Commands{}
	command.AddCommand(NewAddLinkCommand(base))

	t.Run("invalid command", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "add a link"

		actual := command.Run(message)
		assert.False(t, actual)
	})

	t.Run("add link", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "add link example <https://example.com>"

		expected := `[{"actions":[{"name":"","text":"example","style":"default","type":"button","url":"https://example.com"}],"blocks":null}]`
		mocks.AssertSlackJSON(t, slackClient, message, expected)

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("add plain link", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "add link google https://sub.google.com"

		expected := `[{"actions":[{"name":"","text":"google","style":"default","type":"button","url":"https://sub.google.com"}],"blocks":null}]`
		mocks.AssertSlackJSON(t, slackClient, message, expected)

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("add link with quotes", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "add link 'to test google' <https://test.google.com>"

		expected := `[{"actions":[{"name":"","text":"to test google","style":"default","type":"button","url":"https://test.google.com"}],"blocks":null}]`
		mocks.AssertSlackJSON(t, slackClient, message, expected)

		actual := command.Run(message)
		assert.True(t, actual)
	})
}
