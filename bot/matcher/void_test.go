package matcher

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/stretchr/testify/assert"
)

func TestVoid(t *testing.T) {
	t.Run("Match void", func(t *testing.T) {
		subject := NewVoidMatcher()

		message := msg.Message{}
		message.Text = "foo"
		run, match := subject.Match(message)
		assert.Nil(t, match)
		assert.Nil(t, run)
	})
}
