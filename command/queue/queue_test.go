package queue

import (
	"fmt"
	"github.com/innogames/slack-bot/bot/msg"
	"strconv"
	"testing"
	"time"

	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/mocks"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestQueue(t *testing.T) {
	client.InternalMessages = make(chan msg.Message, 2)
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	message := msg.Message{}
	message.User = "testUser1"

	command := bot.Commands{}
	command.AddCommand(NewQueueCommand(base))
	command.AddCommand(NewListCommand(base))

	t.Run("Invalid command", func(t *testing.T) {
		message := msg.Message{}
		actual := command.Run(message)
		assert.False(t, actual)
		assert.Empty(t, client.InternalMessages)
	})

	t.Run("No command running", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "queue reply test"

		slackClient.On("ReplyError", message, fmt.Errorf("you have to call this command when another long running command is already running")).Return(true)
		actual := command.Run(message)
		assert.True(t, actual)
		assert.Empty(t, client.InternalMessages)
	})

	t.Run("No command from other user running", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "queue reply test"

		message2 := msg.Message{}
		message2.User = "testUser2"
		AddRunningCommand(
			message2,
			"",
		)

		slackClient.On("ReplyError", message, fmt.Errorf("you have to call this command when another long running command is already running")).Return(true)
		actual := command.Run(message)
		assert.True(t, actual)
		assert.Empty(t, client.InternalMessages)
	})

	t.Run("Test queue command", func(t *testing.T) {
		now := time.Now()
		message.Timestamp = strconv.Itoa(int(now.Unix()))
		message.Text = "queue reply test"
		done := AddRunningCommand(message, "test")
		msgRef := slack.NewRefToMessage(message.Channel, message.Timestamp)

		slackClient.On("AddReaction", waitIcon, message)
		slackClient.On("AddReaction", doneIcon, message)
		slackClient.On("RemoveReaction", waitIcon, message)

		actual := command.Run(message)
		assert.True(t, actual)
		assert.Empty(t, client.InternalMessages)

		// list queue
		message.Text = "list queue"
		slackClient.On("SendMessage", mock.Anything, "1 queued commands", mock.Anything).Return("")

		slackClient.On("GetReactions", msgRef, slack.NewGetReactionsParameters()).Return(
			[]slack.ItemReaction{
				{Name: "test"},
				{Name: "foo"},
			},
			nil,
		)
		actual = command.Run(message)
		assert.True(t, actual)

		// list queue for current channel
		message.Text = "list queue in channel"
		slackClient.On("SendMessage", mock.Anything, "1 queued commands", mock.Anything).Return("")

		actual = command.Run(message)
		assert.True(t, actual)

		// list queue for other channel
		message.Text = "list queue in channel"
		message.Channel = "C1212121"
		slackClient.On("SendMessage", mock.Anything, "0 queued commands", mock.Anything).Return("")

		actual = command.Run(message)
		assert.True(t, actual)

		done <- true
		time.Sleep(time.Millisecond * 400)

		assert.NotEmpty(t, client.InternalMessages)

		handledEvent := <-client.InternalMessages

		expectedMessage := msg.Message{}
		expectedMessage.Timestamp = message.Timestamp
		expectedMessage.User = "testUser1"
		expectedMessage.Text = "reply test"
		assert.Equal(t, handledEvent, expectedMessage)
	})
}
