package command

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSendMessage(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	command := bot.Commands{}
	command.AddCommand(NewSendMessageCommand(slackClient))

	t.Run("Send to invalid slack user id", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "send message @testuser message"
		message.User = "testUser1"

		actual := command.Run(message)
		assert.Equal(t, false, actual)
	})

	t.Run("Send to user", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "send message <@1234|testuser> message"
		message.User = "testUser1"

		slackClient.On("SendToUser", "1234", "Text from <@testUser1>: message").Return("")
		slackClient.On("SendMessage", message, "I'll send `message` to <@1234|testuser>").Return("")
		actual := command.Run(message)
		assert.Equal(t, true, actual)
	})

	t.Run("Send to channel", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "send message <#JDGS|general> message"
		message.User = "testUser1"

		expectedMessage := msg.Message{}
		expectedMessage.Channel = "JDGS"
		slackClient.On("SendMessage", expectedMessage, "Text from <@testUser1>: message").Return("")
		slackClient.On("SendMessage", message, "I'll send `message` to <#JDGS|general>").Return("")
		actual := command.Run(message)
		assert.Equal(t, true, actual)
	})
}
