package matcher

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/stretchr/testify/assert"
)

func TestWildcard(t *testing.T) {
	t.Run("Match", func(t *testing.T) {
		runner := func(ref msg.Ref, text string) bool {
			return true
		}
		subject := WildcardMatcher(runner)

		message := msg.Message{}
		message.Text = "any"
		run, match := subject.Match(message)

		assert.NotNil(t, match)
		assert.Nil(t, run)
	})

	t.Run("NoMatch", func(t *testing.T) {
		runner := func(ref msg.Ref, text string) bool {
			return false
		}
		subject := WildcardMatcher(runner)

		message := msg.Message{}
		message.Text = "any"
		run, match := subject.Match(message)

		assert.Nil(t, match)
		assert.Nil(t, run)
	})
}
