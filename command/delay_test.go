package command

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/mocks"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDelay(t *testing.T) {
	client.InternalMessages = make(chan slack.MessageEvent, 2)
	slackClient := mocks.SlackClient{}

	command := bot.Commands{}
	command.AddCommand(NewDelayCommand(&slackClient))

	t.Run("Invalid command", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "I have a delay"

		actual := command.Run(event)
		assert.Equal(t, false, actual)
		assert.Empty(t, client.InternalMessages)
	})

	t.Run("Invalid timer", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "delay h1 my command"

		slackClient.On("Reply", event, "Invalid duration: time: invalid duration \"h1\"")
		actual := command.Run(event)
		assert.Equal(t, true, actual)
		assert.Empty(t, client.InternalMessages)
	})

	t.Run("Test timer passed", func(t *testing.T) {
		command := bot.Commands{}
		command.AddCommand(NewDelayCommand(&slackClient))

		event := slack.MessageEvent{}
		event.Text = "delay 20ms my command"

		slackClient.On("Reply", event, "I queued the command `my command` for 20ms. Use `stop timer 0` to stop the timer")
		actual := command.Run(event)
		assert.Equal(t, true, actual)
		assert.Empty(t, client.InternalMessages)

		time.Sleep(time.Millisecond * 250)
		assert.NotEmpty(t, client.InternalMessages)

		handledEvent := <-client.InternalMessages
		expectedEvent := slack.MessageEvent{
			Msg: slack.Msg{
				Text: "my command",
			},
		}

		assert.Equal(t, handledEvent, expectedEvent)
	})

	t.Run("Test quiet option", func(t *testing.T) {
		command := bot.Commands{}
		command.AddCommand(NewDelayCommand(&slackClient))

		event := slack.MessageEvent{}
		event.Text = "delay 20ms quiet my command"

		actual := command.Run(event)
		assert.Equal(t, true, actual)
		assert.Empty(t, client.InternalMessages)

		time.Sleep(time.Millisecond * 100)
		assert.NotEmpty(t, client.InternalMessages)

		handledEvent := <-client.InternalMessages
		expectedEvent := slack.MessageEvent{
			Msg: slack.Msg{
				Text: "my command",
			},
		}

		assert.Equal(t, handledEvent, expectedEvent)
	})

	t.Run("Test stop", func(t *testing.T) {
		command := bot.Commands{}
		command.AddCommand(NewDelayCommand(&slackClient))

		event := slack.MessageEvent{}
		event.Text = "delay 20ms my command"

		slackClient.On("Reply", event, "I queued the command `my command` for 20ms. Use `stop timer 0` to stop the timer")
		actual := command.Run(event)
		assert.Equal(t, true, actual)
		assert.Empty(t, client.InternalMessages)

		event.Text = "stop timer 0"
		slackClient.On("Reply", event, "Stopped timer!")
		actual = command.Run(event)
		assert.Equal(t, true, actual)

		time.Sleep(time.Millisecond * 30)
		assert.Empty(t, client.InternalMessages)

		// now try to stop an invalid timer
		event.Text = "stop timer 5"
		slackClient.On("ReplyError", event, fmt.Errorf("invalid timer")).Return("")
		actual = command.Run(event)
		assert.Equal(t, true, actual)
	})
}
