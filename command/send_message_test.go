package command

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
)

func TestSendMessage(t *testing.T) {
	slackClient := mocks.NewSlackClient(t)
	base := bot.BaseCommand{SlackClient: slackClient}

	command := bot.Commands{}
	command.AddCommand(NewSendMessageCommand(base))

	t.Run("Send to invalid slack user id", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "send message @testuser message"
		message.User = "testUser1"

		actual := command.Run(message)
		assert.False(t, actual)
	})

	t.Run("Send to user", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "send message <@1234|testuser> message"
		message.User = "testUser1"

		slackClient.On("SendToUser", "1234", "Text from <@testUser1>: message").Return("")
		mocks.AssertSlackMessage(slackClient, message, "I'll send `message` to <@1234|testuser>")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Send to channel", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "send message <#JDGS|general> message"
		message.User = "testUser1"

		expectedMessage := msg.Message{}
		expectedMessage.Channel = "JDGS"
		mocks.AssertSlackMessage(slackClient, expectedMessage, "Text from <@testUser1>: message")
		mocks.AssertSlackMessage(slackClient, message, "I'll send `message` to <#JDGS|general>")
		actual := command.Run(message)
		assert.True(t, actual)
	})
}
