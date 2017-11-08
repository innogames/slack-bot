package matcher

import (
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
	"testing"
)

func BenchmarkTextMatcher(b *testing.B) {
	textMatcher := NewTextMatcher("i am the loweredText", testRunner)

	var run Runner
	var result Result

	b.Run("loweredText: no match", func(b *testing.B) {
		event := slack.MessageEvent{}
		event.Text = "i am not the loweredText"

		for i := 0; i < b.N; i++ {
			run, result = textMatcher.Match(event)
		}
		assert.Nil(b, run)
		assert.Equal(b, false, result.Matched())
	})

	b.Run("loweredText: match", func(b *testing.B) {
		event := slack.MessageEvent{}
		event.Text = "i am the loweredText"

		for i := 0; i < b.N; i++ {
			run, result = textMatcher.Match(event)
		}
		assert.NotNil(b, run)
		assert.Equal(b, true, result.Matched())
	})
}
