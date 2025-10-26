package matcher

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/stretchr/testify/assert"
)

func TestWildcard(t *testing.T) {
	t.Run("Match", func(t *testing.T) {
		runner := func(_ msg.Ref, _ string) bool {
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
		runner := func(_ msg.Ref, _ string) bool {
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

func TestWildcardAdvanced(t *testing.T) {
	t.Run("handler execution flow", func(t *testing.T) {
		executed := false
		handledText := ""
		handledChannel := ""
		handledUser := ""

		matcher := WildcardMatcher(func(ref msg.Ref, text string) bool {
			executed = true
			handledText = text
			handledChannel = ref.GetChannel()
			handledUser = ref.GetUser()
			return true // Command was executed, stop processing
		})

		message := msg.Message{}
		message.MessageRef = msg.MessageRef{
			Channel: "C1234567890",
			User:    "U0987654321",
		}
		message.Text = "any command text"

		run, match := matcher.Match(message)
		assert.Nil(t, run)      // Wildcard returns nil for run (stops processing)
		assert.NotNil(t, match) // But returns a match to indicate it handled it
		assert.True(t, executed, "Wildcard handler should have been executed")
		assert.Equal(t, "any command text", handledText)
		assert.Equal(t, "C1234567890", handledChannel)
		assert.Equal(t, "U0987654321", handledUser)
	})

	t.Run("handler returns false continues processing", func(t *testing.T) {
		executed := false

		matcher := WildcardMatcher(func(_ msg.Ref, _ string) bool {
			executed = true
			return false // Command was not handled, continue with other matchers
		})

		message := msg.Message{}
		message.Text = "some text"

		run, match := matcher.Match(message)
		assert.Nil(t, run)
		assert.Nil(t, match) // When handler returns false, both are nil
		assert.True(t, executed, "Wildcard handler should have been executed")
		// Other matchers should continue processing since both are nil
	})

	t.Run("wildcard with empty text", func(t *testing.T) {
		executed := false

		matcher := WildcardMatcher(func(_ msg.Ref, text string) bool {
			executed = true
			assert.Empty(t, text)
			return true
		})

		message := msg.Message{}
		message.Text = ""

		run, match := matcher.Match(message)
		assert.Nil(t, run)
		assert.NotNil(t, match)
		assert.True(t, executed)
	})

	t.Run("wildcard with unicode characters", func(t *testing.T) {
		executed := false
		handledText := ""

		matcher := WildcardMatcher(func(_ msg.Ref, text string) bool {
			executed = true
			handledText = text
			return true
		})

		message := msg.Message{}
		message.Text = "café menu 🚀"

		run, match := matcher.Match(message)
		assert.Nil(t, run)
		assert.NotNil(t, match)
		assert.True(t, executed)
		assert.Equal(t, "café menu 🚀", handledText)
	})
}
