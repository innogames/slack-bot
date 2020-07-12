package command

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/mocks"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestReply(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	reply := NewReplyCommand(slackClient)

	command := bot.Commands{}
	command.AddCommand(reply)

	t.Run("invalid command", func(t *testing.T) {
		texts := []string{
			"i need a reply",
			"replyno",
		}

		for _, text := range texts {
			event := slack.MessageEvent{}
			event.Text = text

			actual := command.Run(event)
			assert.Equal(t, false, actual)
		}
	})

	t.Run("reply without text", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "reply"

		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})

	t.Run("simple reply", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "reply Test"

		slackClient.On("Reply", event, "Test")
		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})

	t.Run("simple reply case sensitive", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "reply Test"

		slackClient.On("SendMessage", event, "Test").Return("")
		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})

	t.Run("comment without text", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "comment"

		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})

	t.Run("comment", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "comment test"
		event.Timestamp = "1234"

		slackClient.On("SendMessage", event, "test", mock.AnythingOfType("slack.MsgOption")).Return("")
		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})
}
