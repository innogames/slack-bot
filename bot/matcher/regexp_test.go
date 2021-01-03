package matcher

import (
	"testing"

	"github.com/innogames/slack-bot/bot/msg"
	"github.com/stretchr/testify/assert"
)

var testRunner = func(match Result, message msg.Message) {}

func TestRegexp(t *testing.T) {
	var matchTest = []struct {
		regexp   string
		input    string
		expected bool
	}{
		{"test", "", false},
		{"test", "notest", false},
		{"test", "ss\ntest", false},
		{"test", "tesT ", false},
		{"test", "TeSt", true},
		{"test", "tesT", true},
		{"test?", "test", true},
		{"test (?P<number>\\d+)", "test 12", true},
	}

	t.Run("Match simple", func(t *testing.T) {
		for _, testCase := range matchTest {
			subject := NewRegexpMatcher(testCase.regexp, testRunner)

			message := msg.Message{}
			message.Text = testCase.input
			run, match := subject.Match(message)
			if testCase.expected {
				assert.NotNil(t, run, testCase.input)
				assert.Equal(t, testCase.input, match.MatchedString())
			} else {
				assert.Nil(t, run, testCase.input)
			}
			assert.Equal(t, testCase.expected, match.Matched())
		}
	})

	t.Run("Get number", func(t *testing.T) {
		subject := NewRegexpMatcher("test (?P<number>\\d+)", testRunner)
		message := msg.Message{}
		message.Text = "test 12"

		run, match := subject.Match(message)
		assert.NotNil(t, run)
		assert.Equal(t, "test 12", match.MatchedString())
		assert.Equal(t, "12", match.GetString("number"))
		assert.Equal(t, 12, match.GetInt("number"))
		assert.Equal(t, true, match.Matched())
	})
}

func BenchmarkRegexpMatcher(b *testing.B) {
	var run Runner
	var result Result

	b.Run("no match", func(b *testing.B) {
		matcher := NewRegexpMatcher("trigger (?P<text>.*)", testRunner)

		message := msg.Message{}
		message.Text = "triggermenot"

		for i := 0; i < b.N; i++ {
			run, result = matcher.Match(message)
		}
		assert.Nil(b, run)
		assert.Equal(b, false, result.Matched())
	})

	b.Run("match", func(b *testing.B) {
		matcher := NewRegexpMatcher("trigger (?P<text>.*)", testRunner)

		message := msg.Message{}
		message.Text = "trigger me"

		for i := 0; i < b.N; i++ {
			run, result = matcher.Match(message)
		}
		assert.NotNil(b, run)
		assert.Equal(b, true, result.Matched())
		assert.Equal(b, "trigger me", result.MatchedString())
	})
}
