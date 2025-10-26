package matcher

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/stretchr/testify/assert"
)

func TestPrefix(t *testing.T) {
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

		{"test", "TeSt", true, ""},
		{"test", "tesT", true, ""},
		{"test", "tesT ", true, ""},
		{"test", "tesT me", true, "me"},
	}

	t.Run("Match simple", func(t *testing.T) {
		for _, testCase := range matchTest {
			subject := NewPrefixMatcher(testCase.prefix, testRunner)

			message := msg.Message{}
			message.Text = testCase.input
			run, match := subject.Match(message)
			if testCase.expected {
				assert.NotNil(t, run, testCase.input)
				assert.Equal(t, testCase.match, match.GetString(util.FullMatch))
			} else {
				assert.Nil(t, run, testCase.input)
				assert.Nil(t, match, testCase.input)
			}
		}
	})

	t.Run("Match simple", func(t *testing.T) {
		subject := NewPrefixMatcher("test", testRunner)

		message := msg.Message{}
		message.Text = "test 15"
		run, match := subject.Match(message)
		assert.NotNil(t, run)
		assert.Equal(t, "15", match.GetString(util.FullMatch))
		assert.Equal(t, 15, match.GetInt(util.FullMatch))
	})
}

func TestPrefixEdgeCases(t *testing.T) {
	t.Run("word boundary detection", func(t *testing.T) {
		testCases := []struct {
			prefix    string
			input     string
			expected  bool
			fullMatch string
		}{
			{"random", "random", true, ""},
			{"random", "random 123", true, "123"},
			{"random", "randomness", false, ""},  // Should not match "randomness"
			{"random", "random\n123", false, ""}, // Should not match across newline
			{"random", "random\t123", false, ""}, // Should not match across tab
			{"test", "testing", false, ""},       // Should not match when prefix is substring
			{"test", "test ", true, ""},          // Should match with trailing space
		}

		for _, testCase := range testCases {
			executed := false
			matcher := NewPrefixMatcher(testCase.prefix, func(match Result, _ msg.Message) {
				executed = true
				if testCase.fullMatch != "" {
					assert.Equal(t, testCase.fullMatch, match.GetString(util.FullMatch))
				}
			})

			message := msg.Message{}
			message.Text = testCase.input

			run, match := matcher.Match(message)
			if testCase.expected {
				assert.NotNil(t, run, "Should match: %s with prefix %s", testCase.input, testCase.prefix)
				assert.NotNil(t, match, "Should have match result")
				if run != nil {
					run(match, message)
				}
				assert.True(t, executed, "Handler should have been executed")
			} else {
				assert.Nil(t, run, "Should not match: %s with prefix %s", testCase.input, testCase.prefix)
				assert.Nil(t, match, "Should not have match result")
			}
		}
	})

	t.Run("case insensitive matching", func(t *testing.T) {
		executed := false
		matcher := NewPrefixMatcher("Hello", func(_ Result, _ msg.Message) {
			executed = true
		})

		message := msg.Message{}
		message.Text = "HELLO world"

		run, match := matcher.Match(message)
		assert.NotNil(t, run)
		assert.NotNil(t, match)

		if run != nil {
			run(match, message)
		}
		assert.True(t, executed)
	})

	t.Run("empty prefix", func(t *testing.T) {
		matcher := NewPrefixMatcher("", testRunner)

		message := msg.Message{}
		message.Text = "any text"

		_, _ = matcher.Match(message)
		// Empty prefix with word boundary logic - this might not match
		// Let's test with text that starts with space
		message.Text = " any text"

		run, match := matcher.Match(message)
		assert.NotNil(t, run) // Empty prefix should match text starting with space
		assert.NotNil(t, match)
	})

	t.Run("unicode characters", func(t *testing.T) {
		executed := false
		matcher := NewPrefixMatcher("café", func(_ Result, _ msg.Message) {
			executed = true
		})

		message := msg.Message{}
		message.Text = "café menu"

		run, match := matcher.Match(message)
		assert.NotNil(t, run)
		assert.NotNil(t, match)

		if run != nil {
			run(match, message)
		}
		assert.True(t, executed)
	})
}
