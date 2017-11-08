package command

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/mocks"
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSendMessage(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	command := bot.Commands{}
	command.AddCommand(NewSendMessageCommand(slackClient))

	t.Run("Send to invalid slack user id", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "send message @testuser message"
		event.User = "testUser1"

		actual := command.Run(event)
		assert.Equal(t, false, actual)
	})

	t.Run("Send to user", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "send message <@1234|testuser> message"
		event.User = "testUser1"

		slackClient.On("SendToUser", "1234", "Message from <@testUser1>: message").Return("")
		slackClient.On("Reply", event, "I'll send `message` to <@1234|testuser>")
		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})

	t.Run("Send to channel", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "send message <#JDGS|general> message"
		event.User = "testUser1"

		slackClient.On("Reply", slack.MessageEvent{Msg: slack.Msg{Channel: "JDGS"}}, "Message from <@testUser1>: message").Return("")
		slackClient.On("Reply", event, "I'll send `message` to <#JDGS|general>")
		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})

}
