package util

import (
	"testing"
	"time"

	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
)

func TestSlack(t *testing.T) {
	t.Run("GetEventKey", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.User = "U1234"
		event.Channel = "C1234"
		event.ThreadTimestamp = "234"

		actual := GetFullEventKey(event)
		assert.Equal(t, "U1234-C1234-234", actual)
		actual = GetShortEventKey(event)
		assert.Equal(t, "U1234-C1234", actual)

		event.ThreadTimestamp = ""
		actual = GetFullEventKey(event)
		assert.Equal(t, "U1234-C1234", actual)
		actual = GetShortEventKey(event)
		assert.Equal(t, "U1234-C1234", actual)
	})

	t.Run("GetEventKey", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Timestamp = "1355517523.000005"

		time.Local, _ = time.LoadLocation("Europe/Berlin")

		actual := GetMessageTime(event)

		expected := "2012-12-14T21:38:43+01:00"
		assert.Equal(t, expected, actual.Format(time.RFC3339))
	})
}
