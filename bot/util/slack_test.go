package util

import (
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
	"testing"
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

}
