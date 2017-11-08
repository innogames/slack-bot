package matcher

import (
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConditional(t *testing.T) {
	t.Run("Match", func(t *testing.T) {
		runner := func(event slack.MessageEvent) bool {
			return true
		}
		subject := NewConditionalMatcher(runner)

		event := slack.MessageEvent{}
		event.Text = "any"
		_, match := subject.Match(event)

		assert.True(t, match.Matched())
	})

	t.Run("NoMatch", func(t *testing.T) {
		runner := func(event slack.MessageEvent) bool {
			return false
		}
		subject := NewConditionalMatcher(runner)

		event := slack.MessageEvent{}
		event.Text = "any"
		_, match := subject.Match(event)

		assert.False(t, match.Matched())
	})
}
