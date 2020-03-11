package command

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/mocks"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestRandom(t *testing.T) {
	slackClient := mocks.SlackClient{}

	command := bot.Commands{}
	command.AddCommand(NewRandomCommand(&slackClient))

	rand.Seed(1) // we want always the same random

	t.Run("invalid command", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "randomness"

		actual := command.Run(event)
		assert.Equal(t, false, actual)
	})

	t.Run("no options should not match", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "random"

		slackClient.On("Reply", event, "You have to pass more arguments")

		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})

	t.Run("pick random with one entry", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "random 1"

		slackClient.On("Reply", event, "1")

		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})

	t.Run("pick random entry", func(t *testing.T) {
		rand.Seed(1) // we want always the same random

		event := slack.MessageEvent{}
		event.Text = "random 1 2 3"

		// seed was chosen to pick the "3" every time
		slackClient.On("Reply", event, "3")

		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})
}
