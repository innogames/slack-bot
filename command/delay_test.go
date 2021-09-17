package command

import (
	"testing"
	"time"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/innogames/slack-bot/v2/command/queue"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestDelay(t *testing.T) {
	client.InternalMessages = make(chan msg.Message, 2)
	slackClient := &mocks.SlackClient{}
	slackClient.On("CanHandleInteractions").Return(true)

	base := bot.BaseCommand{SlackClient: slackClient}

	command := bot.Commands{}
	command.AddCommand(NewDelayCommand(base))

	lock := mocks.LockInternalMessages()
	defer lock.Unlock()

	t.Run("Invalid command", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "I have a delay"

		actual := command.Run(message)
		assert.False(t, actual)
		assert.Empty(t, client.InternalMessages)
	})

	t.Run("Invalid timer", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "delay h1 my command"

		mocks.AssertSlackMessage(slackClient, message, "Invalid duration: time: invalid duration \"h1\"")

		actual := command.Run(message)
		assert.True(t, actual)
		assert.Empty(t, client.InternalMessages)
	})

	t.Run("Test timer passed", func(t *testing.T) {
		command := bot.Commands{}
		command.AddCommand(NewDelayCommand(base))

		message := msg.Message{}
		message.Text = "delay 20ms my command"

		expected := "[{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\"I queued the command `my command` for 20ms. Use `stop timer 0` to stop the timer\"}},{\"type\":\"actions\",\"elements\":[{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Stop timer!\",\"emoji\":true},\"action_id\":\"id\",\"value\":\"stop timer 0\"}]}]"
		mocks.AssertSlackBlocks(t, slackClient, message, expected)

		actual := command.Run(message)
		assert.True(t, actual)
		assert.Empty(t, client.InternalMessages)
		assert.Equal(t, 1, queue.CountCurrentJobs())

		handledEvent := mocks.WaitTillHavingInternalMessage()

		expectedEvent := msg.Message{
			Text: "my command",
		}

		assert.Equal(t, handledEvent, expectedEvent)
	})

	t.Run("Test quiet option", func(t *testing.T) {
		command := bot.Commands{}
		command.AddCommand(NewDelayCommand(base))

		message := msg.Message{}
		message.Text = "delay 20ms quiet my command"

		actual := command.Run(message)
		assert.True(t, actual)
		assert.Empty(t, client.InternalMessages)

		handledEvent := mocks.WaitTillHavingInternalMessage()
		expectedEvent := &slack.MessageEvent{
			Msg: slack.Msg{
				Text: "my command",
			},
		}

		assert.Equal(t, handledEvent, msg.FromSlackEvent(expectedEvent))
	})

	t.Run("Test stop", func(t *testing.T) {
		command := bot.Commands{}
		command.AddCommand(NewDelayCommand(base))

		message := msg.Message{}
		message.Text = "delay 20ms my command"

		expected := "[{\"type\":\"section\",\"text\":{\"type\":\"mrkdwn\",\"text\":\"I queued the command `my command` for 20ms. Use `stop timer 0` to stop the timer\"}},{\"type\":\"actions\",\"elements\":[{\"type\":\"button\",\"text\":{\"type\":\"plain_text\",\"text\":\"Stop timer!\",\"emoji\":true},\"action_id\":\"id\",\"value\":\"stop timer 0\"}]}]"
		mocks.AssertSlackBlocks(t, slackClient, message, expected)

		actual := command.Run(message)
		assert.True(t, actual)
		assert.Empty(t, client.InternalMessages)

		message.Text = "stop timer 0"
		mocks.AssertSlackMessage(slackClient, message, "Stopped timer!")
		actual = command.Run(message)
		assert.True(t, actual)

		time.Sleep(time.Millisecond * 30)
		assert.Empty(t, client.InternalMessages)

		// now try to stop an invalid timer
		message.Text = "stop timer 5"
		mocks.AssertError(slackClient, message, "invalid timer")
		actual = command.Run(message)
		assert.True(t, actual)
	})
}
