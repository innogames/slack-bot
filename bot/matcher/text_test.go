package matcher

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/stretchr/testify/assert"
)

func TestText(t *testing.T) {
	matchTest := []struct {
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
				assert.NotNil(t, match, testCase.input)
			} else {
				assert.Nil(t, run)
				assert.Nil(t, match)
			}
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
		assert.Nil(b, result)
	})

	b.Run("loweredText: match", func(b *testing.B) {
		message := msg.Message{}
		message.Text = "i am the loweredText"

		for i := 0; i < b.N; i++ {
			run, result = textMatcher.Match(message)
		}
		assert.NotNil(b, run)
		assert.NotNil(b, result)
	})
}
