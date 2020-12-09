package command

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestReply(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	reply := NewReplyCommand(base)

	command := bot.Commands{}
	command.AddCommand(reply)

	t.Run("invalid command", func(t *testing.T) {
		texts := []string{
			"i need a reply",
			"replyno",
		}

		for _, text := range texts {
			message := msg.Message{}
			message.Text = text

			actual := command.Run(message)
			assert.False(t, actual)
		}
	})

	t.Run("reply without text", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "reply"

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("simple reply", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "reply Test"

		slackClient.On("SendMessage", message, "Test").Return("")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("simple reply case sensitive", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "reply Test"

		slackClient.On("SendMessage", message, "Test").Return("")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("comment without text", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "comment"

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("comment", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "comment test"
		message.Timestamp = "1234"

		slackClient.On("SendMessage", message, "test", mock.AnythingOfType("slack.MsgOption")).Return("")
		actual := command.Run(message)
		assert.True(t, actual)
	})
}
