package matcher

import (
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestText(t *testing.T) {
	var matchTest = []struct {
		prefix   string
		input    string
		expected bool
		match    string
	}{
		{"test", "", false, ""},
		{"test", "notest", false, ""},
		{"test", "ss\ntest", false, ""},
		{"test", "tesTer", false, ""},
		{"test", "tesT ", false, ""},
		{"test", "tesT me", false, ""},

		{"test", "TeSt", true, "TeSt"},
		{"test", "test", true, "test"},
		{"test", "tesT", true, "tesT"},
	}

	t.Run("Match simple", func(t *testing.T) {
		for _, testCase := range matchTest {
			subject := NewTextMatcher(testCase.prefix, testRunner)

			message := msg.Message{}
			message.Text = testCase.input
			run, match := subject.Match(message)
			if testCase.expected {
				assert.NotNil(t, run, testCase.input)
				assert.Equal(t, testCase.match, match.MatchedString())
			} else {
				assert.Nil(t, run, testCase.input)
			}
			assert.Equal(t, testCase.expected, match.Matched())
		}
	})
}

func BenchmarkTextMatcher(b *testing.B) {
	textMatcher := NewTextMatcher("i am the loweredText", testRunner)

	var run Runner
	var result Result

	b.Run("loweredText: no match", func(b *testing.B) {
		message := msg.Message{}
		message.Text = "i am not the loweredText"

		for i := 0; i < b.N; i++ {
			run, result = textMatcher.Match(message)
		}
		assert.Nil(b, run)
		assert.False(b, result.Matched())
	})

	b.Run("loweredText: match", func(b *testing.B) {
		message := msg.Message{}
		message.Text = "i am the loweredText"

		for i := 0; i < b.N; i++ {
			run, result = textMatcher.Match(message)
		}
		assert.NotNil(b, run)
		assert.True(b, result.Matched())
	})
}
