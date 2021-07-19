package command

import (
	"fmt"
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestRetry(t *testing.T) {
	lock := mocks.LockInternalMessages()
	defer lock.Unlock()

	client.InternalMessages = make(chan msg.Message, 2)
	slackClient := &mocks.SlackClient{}
	cfg := &config.Config{}
	base := bot.BaseCommand{SlackClient: slackClient}

	retry := bot.Commands{}
	retry.AddCommand(NewRetryCommand(base, cfg))

	t.Run("Ignore internal messages", func(t *testing.T) {
		message := msg.Message{}
		message.User = "testUser1"
		message.Text = "i'm a submessage"
		message.InternalMessage = true

		actual := retry.Run(message)
		assert.False(t, actual)
		assert.Empty(t, client.InternalMessages)
	})

	t.Run("Full test", func(t *testing.T) {
		// no retry available
		message := msg.Message{}
		message.User = "testUser1"
		message.Text = "retry"

		mocks.AssertSlackMessage(slackClient, message, "Sorry, no history found.")

		actual := retry.Run(message)
		assert.True(t, actual)
		assert.Empty(t, client.InternalMessages)

		// send any other command
		message = msg.Message{}
		message.User = "testUser1"
		message.Text = "magic command"
		actual = retry.Run(message)
		assert.False(t, actual)
		assert.Empty(t, client.InternalMessages)

		// retry -> "magic command"
		message2 := msg.Message{}
		message2.User = "testUser1"
		message2.Text = "retry"

		mocks.AssertSlackMessage(slackClient, message2, "Executing command: magic command")

		actual = retry.Run(message2)
		assert.True(t, actual)
		assert.NotEmpty(t, client.InternalMessages)

		handledEvent := <-client.InternalMessages
		assert.Equal(t, handledEvent, message)
	})

	t.Run("Run with other user", func(t *testing.T) {
		message := msg.Message{}
		message.User = "testUser1"
		message.Text = "magic command"

		actual := retry.Run(message)

		assert.False(t, actual)
		assert.Empty(t, client.InternalMessages)

		message2 := msg.Message{}
		message2.User = "testUser2"
		message2.Text = "retry"

		mocks.AssertSlackMessage(slackClient, message2, "Sorry, no history found.")

		actual = retry.Run(message2)

		assert.True(t, actual)
		assert.Empty(t, client.InternalMessages)
	})

	t.Run("RetryMessage with error", func(t *testing.T) {
		message := msg.Message{}
		message.User = "testUser1"
		message.Channel = "myChan"
		message.Text = "<https://foe-workshop.slack.com/archives/D0183HUURA9/p1607971366001000>"

		err := fmt.Errorf("bad")
		slackClient.On("GetConversationHistory", &slack.GetConversationHistoryParameters{ChannelID: "D0183HUURA9", Inclusive: true, Latest: "1607971366.001000", Limit: 1}).
			Once().
			Return(nil, err)

		mocks.AssertError(slackClient, message, fmt.Errorf("can't load original message: %w", err))
		actual := retry.Run(message)

		assert.True(t, actual)
	})

	t.Run("RetryMessage with different user", func(t *testing.T) {
		message := msg.Message{}
		message.User = "testUser1"
		message.Channel = "myChan"
		message.Text = "<https://foe-workshop.slack.com/archives/D0183HUURA9/p1607971366001000>"

		history := slack.Message{
			Msg: slack.Msg{
				Text: "reply foo",
				User: "bar",
			},
		}

		slackClient.On(
			"GetConversationHistory",
			&slack.GetConversationHistoryParameters{ChannelID: "D0183HUURA9", Inclusive: true, Latest: "1607971366.001000", Limit: 1},
		).Once().Return(&slack.GetConversationHistoryResponse{Messages: []slack.Message{history}}, nil)

		mocks.AssertSlackMessage(slackClient, message, "this is not your message")
		actual := retry.Run(message)

		assert.True(t, actual)
	})

	t.Run("RetryMessage", func(t *testing.T) {
		message := msg.Message{}
		message.User = "testUser1"
		message.Channel = "myChan"
		message.Text = "<https://foe-workshop.slack.com/archives/D0183HUURA9/p1607971366001001>"

		history := slack.Message{
			Msg: slack.Msg{
				Text: "reply foo",
				User: "testUser1",
			},
		}

		slackClient.On(
			"GetConversationHistory",
			&slack.GetConversationHistoryParameters{ChannelID: "D0183HUURA9", Inclusive: true, Latest: "1607971366.001001", Limit: 1},
		).Return(&slack.GetConversationHistoryResponse{Messages: []slack.Message{history}}, nil)
		mocks.AssertReaction(slackClient, "✅", message)
		mocks.AssertSlackMessage(slackClient, message, "this is not your message")

		actual := retry.Run(message)

		assert.True(t, actual)
		assert.NotEmpty(t, client.InternalMessages)

		newMessage := <-client.InternalMessages

		expected := msg.Message{
			MessageRef: msg.MessageRef{
				Channel: "D0183HUURA9",
				User:    history.User,
			},
			Text: "reply foo",
		}
		assert.Equal(t, expected, newMessage)
	})
}
