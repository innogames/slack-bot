package queue

import (
	"strconv"
	"testing"
	"time"

	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/mocks"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
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

	lock := mocks.LockInternalMessages()
	defer lock.Unlock()

	t.Run("Invalid command", func(t *testing.T) {
		message := msg.Message{}
		actual := command.Run(message)
		assert.False(t, actual)
		assert.Empty(t, client.InternalMessages)
	})

	t.Run("No command running", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "queue reply test"

		mocks.AssertError(slackClient, message, "you have to call this command when another long running command is already running")
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

		mocks.AssertError(slackClient, message, "you have to call this command when another long running command is already running")
		actual := command.Run(message)
		assert.True(t, actual)
		assert.Empty(t, client.InternalMessages)
	})

	t.Run("Test queue command", func(t *testing.T) {
		now := time.Now()
		message.Timestamp = strconv.Itoa(int(now.Unix()))
		message.Text = "queue reply test"
		runningCommand := AddRunningCommand(message, "test")
		msgRef := slack.NewRefToMessage(message.Channel, message.Timestamp)

		mocks.AssertReaction(slackClient, waitIcon, message)
		mocks.AssertReaction(slackClient, doneIcon, message)
		mocks.AssertRemoveReaction(slackClient, waitIcon, message)

		actual := command.Run(message)
		assert.True(t, actual)
		assert.Empty(t, client.InternalMessages)

		// list queue
		message.Text = "list queue"
		mocks.AssertReaction(slackClient, processingReaction, message)
		mocks.AssertRemoveReaction(slackClient, processingReaction, message)
		mocks.AssertContainsSlackBlocks(t, slackClient, message, client.GetTextBlock("*1 queued commands*"))

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
		mocks.AssertReaction(slackClient, processingReaction, message)
		mocks.AssertRemoveReaction(slackClient, processingReaction, message)
		mocks.AssertContainsSlackBlocks(t, slackClient, message, client.GetTextBlock("*1 queued commands*"))

		actual = command.Run(message)
		assert.True(t, actual)

		// list queue for other channel
		message.Text = "list queue in channel"
		message.Channel = "C1212121"
		mocks.AssertReaction(slackClient, processingReaction, message)
		mocks.AssertRemoveReaction(slackClient, processingReaction, message)
		mocks.AssertContainsSlackBlocks(t, slackClient, message, client.GetTextBlock("*0 queued commands*"))

		actual = command.Run(message)
		assert.True(t, actual)

		runningCommand.Done()

		mocks.WaitTillHavingInternalMessage()

		handledEvent := <-client.InternalMessages

		expectedMessage := msg.Message{}
		expectedMessage.Timestamp = message.Timestamp
		expectedMessage.User = "testUser1"
		expectedMessage.Text = "reply test"
		assert.Equal(t, handledEvent, expectedMessage)
	})

	t.Run("Test refresh queue command", func(t *testing.T) {
		message.Text = "list queue"
		message.UpdatedMessage = true

		mocks.AssertReaction(slackClient, processingReaction, message)
		mocks.AssertRemoveReaction(slackClient, processingReaction, message)
		mocks.AssertContainsSlackBlocks(t, slackClient, message, client.GetTextBlock("*0 queued commands*"))

		actual := command.Run(message)
		assert.True(t, actual)
	})
}

func TestFallbackQueue(t *testing.T) {
	client.InternalMessages = make(chan msg.Message, 2)
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	message := msg.Message{}
	message.User = "testUser1"

	// this command should get executed on next startup..ot when we initialize "NewQueueCommand
	runningCommand := AddRunningCommand(message, "reply yep")

	command := bot.Commands{}
	command.AddCommand(NewQueueCommand(base))

	handledEvent := <-client.InternalMessages

	expectedMessage := msg.Message{}
	expectedMessage.Timestamp = message.Timestamp
	expectedMessage.User = message.User
	expectedMessage.Text = "reply yep"
	assert.Equal(t, handledEvent, expectedMessage)

	runningCommand.Done()
}
