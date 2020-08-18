package matcher

import (
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"testing"
)

var testRunner = func(match Result, event slack.MessageEvent) {}

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

			event := slack.MessageEvent{}
			event.Text = testCase.input
			run, match := subject.Match(event)
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

		event := slack.MessageEvent{}
		event.Text = "test 12"
		run, match := subject.Match(event)
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

		event := slack.MessageEvent{}
		event.Text = "triggermenot"

		for i := 0; i < b.N; i++ {
			run, result = matcher.Match(event)
		}
		assert.Nil(b, run)
		assert.Equal(b, false, result.Matched())
	})

	b.Run("match", func(b *testing.B) {
		matcher := NewRegexpMatcher("trigger (?P<text>.*)", testRunner)

		event := slack.MessageEvent{}
		event.Text = "trigger me"

		for i := 0; i < b.N; i++ {
			run, result = matcher.Match(event)
		}
		assert.NotNil(b, run)
		assert.Equal(b, true, result.Matched())
	})
}
