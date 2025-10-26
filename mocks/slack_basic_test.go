package mocks

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSlackClientBasic(t *testing.T) {
	t.Run("SendMessage basic test", func(t *testing.T) {
		slackClient := NewSlackClient(t)

		message := msg.Message{
			MessageRef: msg.MessageRef{
				Channel: "C1234567890",
				User:    "U1234567890",
			},
		}

		// Test successful message send
		slackClient.On("SendMessage", message, "test message").Return("C1234567890.1234567890")

		result := slackClient.SendMessage(message, "test message")
		assert.Equal(t, "C1234567890.1234567890", result)

		slackClient.AssertExpectations(t)
	})

	t.Run("AddReaction basic test", func(t *testing.T) {
		slackClient := NewSlackClient(t)

		message := msg.Message{
			MessageRef: msg.MessageRef{
				Channel: "C1234567890",
				User:    "U1234567890",
			},
		}

		// Test successful reaction add
		slackClient.On("AddReaction", mock.AnythingOfType("util.Reaction"), mock.AnythingOfType("msg.Message")).Return(nil)

		slackClient.AddReaction(util.Reaction(":coffee:"), message)

		slackClient.AssertExpectations(t)
	})
}
