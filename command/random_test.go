package command

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
)

func TestRandom(t *testing.T) {
	slackClient := mocks.NewSlackClient(t)
	base := bot.BaseCommand{SlackClient: slackClient}

	randomCommand := NewRandomCommand(base).(*randomCommand)
	randomCommand.random = mocks.NewPseudoRandom() // we want always the same random

	command := bot.Commands{}
	command.AddCommand(randomCommand)

	t.Run("invalid command", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "randomness"

		actual := command.Run(message)
		assert.False(t, actual)
	})

	t.Run("no options should not match", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "random"

		mocks.AssertSlackMessage(slackClient, message, "You have to pass more arguments")

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("pick random with one entry", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "random 1"

		mocks.AssertSlackMessage(slackClient, message, "1")

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("pick random entry", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "random 4 5 6 7"

		// seed was chosen to pick the "6" every time
		mocks.AssertSlackMessage(slackClient, message, "6")
		actual := command.Run(message)
		assert.True(t, actual)

		// second time the 4 is used
		mocks.AssertSlackMessage(slackClient, message, "4")
		actual = command.Run(message)
		assert.True(t, actual)
	})
}
