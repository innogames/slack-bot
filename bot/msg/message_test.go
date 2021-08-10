package msg

import (
	"testing"
	"time"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestMessage(t *testing.T) {
	t.Run("Slack event to Message", func(t *testing.T) {
		event := &slack.MessageEvent{
			Msg: slack.Msg{
				Text:      "foo",
				Timestamp: "12344",
				User:      "U123",
			},
		}
		actual := FromSlackEvent(event)

		expected := Message{
			Text: "foo",
			MessageRef: MessageRef{
				User:            "U123",
				Timestamp:       "12344",
				InternalMessage: false,
			},
		}

		assert.Equal(t, expected, actual)
		assert.Equal(t, false, actual.IsInternalMessage())
	})

	t.Run("Get message time", func(t *testing.T) {
		msg := Message{}
		msg.Timestamp = "1355517523.000000"

		time.Local, _ = time.LoadLocation("Europe/Berlin")

		actual := msg.GetTime()

		expected := "2012-12-14T21:38:43+01:00"
		assert.Equal(t, expected, actual.Format(time.RFC3339Nano))
	})

	t.Run("Get message time with micro", func(t *testing.T) {
		msg := Message{}
		msg.Timestamp = "1355517523.000005"

		time.Local, _ = time.LoadLocation("Europe/Berlin")

		actual := msg.GetTime()

		expected := "2012-12-14T21:38:43.000005+01:00"
		assert.Equal(t, expected, actual.Format(time.RFC3339Nano))
	})

	t.Run("Get Key", func(t *testing.T) {
		message := Message{
			Text: "foo",
			MessageRef: MessageRef{
				Channel:         "chan",
				User:            "U123",
				Timestamp:       "12344",
				InternalMessage: true,
			},
		}
		assert.Equal(t, "U123-chan", message.GetUniqueKey())
	})

	t.Run("Get MessageRef", func(t *testing.T) {
		message := Message{
			Text: "foo",
			MessageRef: MessageRef{
				Channel:         "chan",
				User:            "U123",
				Timestamp:       "12344.111",
				InternalMessage: true,
			},
		}
		expected := slack.ItemRef{
			Channel:   "chan",
			Timestamp: "12344.111",
		}
		assert.Equal(t, expected, message.GetMessageRef())
	})
}
