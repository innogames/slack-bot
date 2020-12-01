package command

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/mocks"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDelay(t *testing.T) {
	client.InternalMessages = make(chan msg.Message, 2)
	slackClient := mocks.SlackClient{}

	command := bot.Commands{}
	command.AddCommand(NewDelayCommand(&slackClient))

	t.Run("Invalid command", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "I have a delay"

		actual := command.Run(message)
		assert.Equal(t, false, actual)
		assert.Empty(t, client.InternalMessages)
	})

	t.Run("Invalid timer", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "delay h1 my command"

		slackClient.On("SendMessage", message, "Invalid duration: time: invalid duration \"h1\"").Return("")
		actual := command.Run(message)
		assert.Equal(t, true, actual)
		assert.Empty(t, client.InternalMessages)
	})

	t.Run("Test timer passed", func(t *testing.T) {
		command := bot.Commands{}
		command.AddCommand(NewDelayCommand(&slackClient))

		message := msg.Message{}
		message.Text = "delay 20ms my command"

		slackClient.On("SendMessage", message, "I queued the command `my command` for 20ms. Use `stop timer 0` to stop the timer").Return("")
		actual := command.Run(message)
		assert.Equal(t, true, actual)
		assert.Empty(t, client.InternalMessages)

		time.Sleep(time.Millisecond * 250)
		assert.NotEmpty(t, client.InternalMessages)

		handledEvent := <-client.InternalMessages
		expectedEvent := msg.Message{
			Text: "my command",
		}

		assert.Equal(t, handledEvent, expectedEvent)
	})

	t.Run("Test quiet option", func(t *testing.T) {
		command := bot.Commands{}
		command.AddCommand(NewDelayCommand(&slackClient))

		message := msg.Message{}
		message.Text = "delay 20ms quiet my command"

		actual := command.Run(message)
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

		assert.Equal(t, handledEvent, msg.FromSlackEvent(expectedEvent))
	})

	t.Run("Test stop", func(t *testing.T) {
		command := bot.Commands{}
		command.AddCommand(NewDelayCommand(&slackClient))

		message := msg.Message{}
		message.Text = "delay 20ms my command"

		slackClient.On("SendMessage", message, "I queued the command `my command` for 20ms. Use `stop timer 0` to stop the timer").Return("")
		actual := command.Run(message)
		assert.Equal(t, true, actual)
		assert.Empty(t, client.InternalMessages)

		message.Text = "stop timer 0"
		slackClient.On("SendMessage", message, "Stopped timer!").Return("")
		actual = command.Run(message)
		assert.Equal(t, true, actual)

		time.Sleep(time.Millisecond * 30)
		assert.Empty(t, client.InternalMessages)

		// now try to stop an invalid timer
		message.Text = "stop timer 5"
		slackClient.On("ReplyError", message, fmt.Errorf("invalid timer"))
		actual = command.Run(message)
		assert.Equal(t, true, actual)
	})
}
