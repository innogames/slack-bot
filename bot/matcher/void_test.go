package matcher

import (
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVoid(t *testing.T) {
	t.Run("Match void", func(t *testing.T) {
		subject := NewVoidMatcher()

		event := slack.MessageEvent{}
		event.Text = "foo"
		run, match := subject.Match(event)
		assert.False(t, match.Matched())
		assert.Nil(t, run)
	})
}
