package command

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRetry(t *testing.T) {
	client.InternalMessages = make(chan msg.Message, 2)
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	retry := bot.Commands{}
	retry.AddCommand(NewRetryCommand(base))

	t.Run("Ignore internal messages", func(t *testing.T) {
		message := msg.Message{}
		message.User = "testUser1"
		message.Text = "i'm a submessage"
		message.InternalMessage = true

		actual := retry.Run(message)
		assert.False(t, actual)
		assert.Empty(t, client.InternalMessages)
	})

	t.Run("Full test", func(t *testing.T) {
		// no retry available
		message := msg.Message{}
		message.User = "testUser1"
		message.Text = "retry"
		slackClient.On("SendMessage", message, "Sorry, no history found.").Return("")
		actual := retry.Run(message)
		assert.True(t, actual)
		assert.Empty(t, client.InternalMessages)

		// send any other command
		message = msg.Message{}
		message.User = "testUser1"
		message.Text = "magic command"
		actual = retry.Run(message)
		assert.False(t, actual)
		assert.Empty(t, client.InternalMessages)

		// retry -> "magic command"
		message2 := msg.Message{}
		message2.User = "testUser1"
		message2.Text = "retry"
		slackClient.On("SendMessage", message2, "Executing command: magic command").Return("")
		actual = retry.Run(message2)
		assert.True(t, actual)
		assert.NotEmpty(t, client.InternalMessages)

		handledEvent := <-client.InternalMessages
		assert.Equal(t, handledEvent, message)
	})

	t.Run("With with other user", func(t *testing.T) {
		message := msg.Message{}
		message.User = "testUser1"
		message.Text = "magic command"

		actual := retry.Run(message)

		assert.False(t, actual)
		assert.Empty(t, client.InternalMessages)

		message2 := msg.Message{}
		message2.User = "testUser2"
		message2.Text = "retry"

		slackClient.On("SendMessage", message2, "Sorry, no history found.").Return("")

		actual = retry.Run(message2)

		assert.True(t, actual)
		assert.Empty(t, client.InternalMessages)
	})
}
