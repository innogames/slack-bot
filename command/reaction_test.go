package command

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/mocks"
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReaction(t *testing.T) {
	slackClient := mocks.SlackClient{}
	reaction := NewReactionCommand(&slackClient)

	command := bot.Commands{}
	command.AddCommand(reaction)

	t.Run("invalid command", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "i need a reaction"

		actual := command.Run(event)
		assert.Equal(t, false, actual)
	})

	t.Run("add reaction", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "add reaction :test:"
		event.Channel = "chan"
		event.Timestamp = "time"

		msgRef := slack.NewRefToMessage(event.Channel, event.Timestamp)

		slackClient.On("AddReaction", "test", msgRef)
		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})

	t.Run("remove reaction", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "remove reaction :test:"
		event.Channel = "chan"
		event.Timestamp = "time"

		msgRef := slack.NewRefToMessage(event.Channel, event.Timestamp)

		slackClient.On("RemoveReaction", "test", msgRef)
		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})
}
