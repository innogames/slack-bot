package msg

import (
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMessage(t *testing.T) {
	t.Run("Slack event to Message", func(t *testing.T) {
		event := slack.MessageEvent{
			Msg: slack.Msg{
				Text:      "foo",
				Timestamp: "12344",
				User:      "U123",
			},
		}
		actual := FromSlackEvent(event)

		expected := Message{
			Text:            "foo",
			User:            "U123",
			Timestamp:       "12344",
			InternalMessage: false,
		}

		assert.Equal(t, expected, actual)
		assert.Equal(t, false, actual.IsInternalMessage())
	})

	t.Run("Message to Slack event", func(t *testing.T) {
		message := Message{
			Text:            "foo",
			User:            "U123",
			Timestamp:       "12344",
			InternalMessage: true,
		}
		assert.Equal(t, true, message.IsInternalMessage())

		expected := slack.MessageEvent{
			Msg: slack.Msg{
				Text:      "foo",
				Timestamp: "12344",
				User:      "U123",
				SubType:   typeInternal,
			},
		}
		actual := message.ToSlackEvent()

		assert.Equal(t, expected, actual)
	})
}
