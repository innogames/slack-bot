package matcher

import (
	"github.com/nlopes/slack"
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
