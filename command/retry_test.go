package command

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/mocks"
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRetry(t *testing.T) {
	client.InternalMessages = make(chan slack.MessageEvent, 2)
	slackClient := mocks.SlackClient{}

	retry := bot.Commands{}
	retry.AddCommand(NewRetryCommand(&slackClient))

	t.Run("Ignore internal messages", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.User = "testUser1"
		event.Text = "i'm a submessage"
		event.SubType = bot.TypeInternal

		actual := retry.Run(event)
		assert.Equal(t, false, actual)
		assert.Empty(t, client.InternalMessages)
	})

	t.Run("Full test", func(t *testing.T) {
		// no retry available
		event := slack.MessageEvent{}
		event.User = "testUser1"
		event.Text = "retry"
		slackClient.On("Reply", event, "Sorry, no history found.")
		actual := retry.Run(event)
		assert.Equal(t, true, actual)
		assert.Empty(t, client.InternalMessages)

		// send any other command
		event = slack.MessageEvent{}
		event.User = "testUser1"
		event.Text = "magic command"
		actual = retry.Run(event)
		assert.Equal(t, false, actual)
		assert.Empty(t, client.InternalMessages)

		// retry -> "magic command"
		event2 := slack.MessageEvent{}
		event2.User = "testUser1"
		event2.Text = "retry"
		slackClient.On("Reply", event2, "Executing command: magic command")
		actual = retry.Run(event2)
		assert.Equal(t, true, actual)
		assert.NotEmpty(t, client.InternalMessages)

		handledEvent := <-client.InternalMessages
		assert.Equal(t, handledEvent, event)
	})

	t.Run("With with other user", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.User = "testUser1"
		event.Text = "magic command"

		actual := retry.Run(event)

		assert.Equal(t, false, actual)
		assert.Empty(t, client.InternalMessages)

		event2 := slack.MessageEvent{}
		event2.User = "testUser2"
		event2.Text = "retry"

		slackClient.On("Reply", event2, "Sorry, no history found.")

		actual = retry.Run(event2)

		assert.Equal(t, true, actual)
		assert.Empty(t, client.InternalMessages)
	})
}
