package command

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRequest(t *testing.T) {
	slackClient := mocks.NewSlackClient(t)
	base := bot.BaseCommand{SlackClient: slackClient}

	reaction := NewRequestCommand(base)

	command := bot.Commands{}
	command.AddCommand(reaction)

	t.Run("invalid command", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "i need a reaction"

		actual := command.Run(message)
		assert.False(t, actual)
	})

	t.Run("without URL", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "request --method=GET"

		mocks.AssertError(slackClient, message, "please provide a valid url")

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("invalid url schema", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "request --url=://example"

		mocks.AssertError(slackClient, message, "invalid request: parse \"://example\": missing protocol scheme")

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("invalid destination", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "request --method=GET --url=https://127.111.111.111"

		mocks.AssertReaction(slackClient, "❌", message)
		slackClient.On("ReplyError", message, mock.Anything).Once().Return("")

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("test 200", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "request --url=https://httpstat.us/200"

		mocks.AssertReaction(slackClient, "white_check_mark", message)

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("test 404", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "request --url=https://httpstat.us/404"

		mocks.AssertReaction(slackClient, "❌", message)
		mocks.AssertError(slackClient, message, "request failed with status 404 Not Found")

		actual := command.Run(message)
		assert.True(t, actual)
	})
}
