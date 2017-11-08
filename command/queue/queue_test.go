package queue

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/storage"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/mocks"
	"github.com/nlopes/slack"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestQueue(t *testing.T) {
	client.InternalMessages = make(chan slack.MessageEvent, 2)
	slackClient := &mocks.SlackClient{}

	after := storage.MockStorage()
	defer after()

	event := slack.MessageEvent{}
	event.User = "testUser1"

	logger, _ := test.NewNullLogger()

	command := bot.Commands{}
	command.AddCommand(NewQueueCommand(slackClient, logger))

	t.Run("Invalid command", func(t *testing.T) {
		event.Text = "I have a queuedCommand"
		actual := command.Run(event)
		assert.Equal(t, false, actual)
		assert.Empty(t, client.InternalMessages)
	})

	t.Run("No command running", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "queue reply test"

		slackClient.On("ReplyError", event, fmt.Errorf("You have to call this command when another long running command is already running")).Return(true)
		actual := command.Run(event)
		assert.Equal(t, true, actual)
		assert.Empty(t, client.InternalMessages)
	})

	t.Run("No command from other user running", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "queue reply test"

		AddRunningCommand(slack.MessageEvent{
			Msg: slack.Msg{
				User: "testUser2",
			}},
			"",
		)

		slackClient.On("ReplyError", event, fmt.Errorf("You have to call this command when another long running command is already running")).Return(true)
		actual := command.Run(event)
		assert.Equal(t, true, actual)
		assert.Empty(t, client.InternalMessages)
	})

	t.Run("Test queue command", func(t *testing.T) {
		done := AddRunningCommand(event, "")

		msgRef := slack.NewRefToMessage(event.Channel, event.Timestamp)

		slackClient.On("AddReaction", waitIcon, msgRef)
		slackClient.On("AddReaction", doneIcon, msgRef)
		slackClient.On("RemoveReaction", waitIcon, msgRef)

		event.Text = "queue reply test"
		actual := command.Run(event)
		assert.Equal(t, true, actual)

		assert.Empty(t, client.InternalMessages)

		// list check list command

		done <- true
		time.Sleep(time.Millisecond * 300)

		assert.NotEmpty(t, client.InternalMessages)

		handledEvent := <-client.InternalMessages
		assert.Equal(t, handledEvent, slack.MessageEvent{
			Msg: slack.Msg{
				User: "testUser1",
				Text: "reply test",
			},
		})
	})
}
