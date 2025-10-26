package matcher

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/stretchr/testify/assert"
)

var testRunner = func(_ Result, _ msg.Message) {}

func TestRegexp(t *testing.T) {
	matchTest := []struct {
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
				assert.NotNil(t, match, testCase.input)
			} else {
				assert.Nil(t, run, testCase.input)
				assert.Nil(t, match, testCase.input)
			}
		}
	})

	t.Run("Get number", func(t *testing.T) {
		subject := NewRegexpMatcher("test (?P<number>\\d+)", testRunner)
		message := msg.Message{}
		message.Text = "test 12"

		run, match := subject.Match(message)
		assert.NotNil(t, run)
		assert.Equal(t, "12", match.GetString("number"))
		assert.Empty(t, match.GetString("number_invalid"))
		assert.Equal(t, 12, match.GetInt("number"))
		assert.Equal(t, 0, match.GetInt("number_invalid"))
		assert.True(t, match.Has("number"))
		assert.False(t, match.Has("invalid"))
	})
}

func BenchmarkRegexpMatcher(b *testing.B) {
	var run Runner
	var result Result

	b.Run("no match", func(b *testing.B) {
		matcher := NewRegexpMatcher("trigger (?P<text>.*)", testRunner)

		message := msg.Message{}
		message.Text = "triggermenot"

		for range b.N {
			run, result = matcher.Match(message)
		}
		assert.Nil(b, run)
		assert.Nil(b, result)
	})

	b.Run("match", func(b *testing.B) {
		matcher := NewRegexpMatcher("trigger (?P<text>.*)", testRunner)

		message := msg.Message{}
		message.Text = "trigger me"

		for range b.N {
			run, result = matcher.Match(message)
		}
		assert.NotNil(b, run)
		assert.NotNil(b, matcher)
		assert.Equal(b, "me", result.GetString("text"))
	})
}

func TestRegexpComplexPatterns(t *testing.T) {
	t.Run("complex named groups", func(t *testing.T) {
		executed := false
		matcher := NewRegexpMatcher(`pool lock\s+(?P<resource>\w+)(\s+reason\s+(?P<reason>.+))?`, func(match Result, message msg.Message) {
			executed = true
			assert.Equal(t, "server1", match.GetString("resource"))
			assert.Equal(t, "for testing", match.GetString("reason"))
			assert.Empty(t, match.GetString("nonexistent"))
		})

		message := msg.Message{}
		message.Text = "pool lock server1 reason for testing"

		run, match := matcher.Match(message)
		assert.NotNil(t, run)
		assert.NotNil(t, match)

		if run != nil {
			run(match, message)
		}
		assert.True(t, executed)
	})

	t.Run("optional groups", func(t *testing.T) {
		executed := false
		matcher := NewRegexpMatcher(`export channel\s+#?(?P<channel>[\w\-]+)\s+as\s+(?P<format>\w+)`, func(match Result, message msg.Message) {
			executed = true
			assert.Equal(t, "general", match.GetString("channel"))
			assert.Equal(t, "csv", match.GetString("format"))
		})

		message := msg.Message{}
		message.Text = "export channel #general as csv"

		run, match := matcher.Match(message)
		assert.NotNil(t, run)
		assert.NotNil(t, match)

		if run != nil {
			run(match, message)
		}
		assert.True(t, executed)
	})

	t.Run("multiple matches in group", func(t *testing.T) {
		executed := false
		matcher := NewRegexpMatcher(`jira\s+(?P<action>create|update|delete)\s+(?P<project>\w+)-(?P<ticket>\d+)`, func(match Result, message msg.Message) {
			executed = true
			assert.Equal(t, "create", match.GetString("action"))
			assert.Equal(t, "PROJ", match.GetString("project"))
			assert.Equal(t, "123", match.GetString("ticket"))
		})

		message := msg.Message{}
		message.Text = "jira create PROJ-123"

		run, match := matcher.Match(message)
		assert.NotNil(t, run)
		assert.NotNil(t, match)

		if run != nil {
			run(match, message)
		}
		assert.True(t, executed)
	})

	t.Run("case insensitive matching", func(t *testing.T) {
		executed := false
		matcher := NewRegexpMatcher("hello world", func(match Result, message msg.Message) {
			executed = true
		})

		message := msg.Message{}
		message.Text = "HELLO WORLD"

		run, match := matcher.Match(message)
		assert.NotNil(t, run)
		assert.NotNil(t, match)

		if run != nil {
			run(match, message)
		}
		assert.True(t, executed)
	})

	t.Run("no match scenarios", func(t *testing.T) {
		testCases := []string{
			"",
			"nomatch",
			"trigger\nnot",
			"trigger extra words",
		}

		matcher := NewRegexpMatcher("trigger", testRunner)

		for _, testCase := range testCases {
			message := msg.Message{}
			message.Text = testCase

			run, match := matcher.Match(message)
			assert.Nil(t, run, "Should not match: %s", testCase)
			assert.Nil(t, match, "Should not match: %s", testCase)
		}
	})
}

func TestRegexpEdgeCases(t *testing.T) {
	t.Run("regex with special characters", func(t *testing.T) {
		executed := false
		matcher := NewRegexpMatcher(`send\s+message\s+to\s+<@(?P<user>\w+)\|(?P<name>[\w\s]+)>`, func(match Result, message msg.Message) {
			executed = true
			assert.Equal(t, "U1234567890", match.GetString("user"))
			assert.Equal(t, "John Doe", match.GetString("name"))
		})

		message := msg.Message{}
		message.Text = "send message to <@U1234567890|John Doe>"

		run, match := matcher.Match(message)
		assert.NotNil(t, run)
		assert.NotNil(t, match)

		if run != nil {
			run(match, message)
		}
		assert.True(t, executed)
	})

	t.Run("numeric parsing", func(t *testing.T) {
		executed := false
		matcher := NewRegexpMatcher(`set\s+priority\s+(?P<priority>\d+)`, func(match Result, message msg.Message) {
			executed = true
			assert.Equal(t, 5, match.GetInt("priority"))
			assert.Equal(t, 0, match.GetInt("nonexistent"))
		})

		message := msg.Message{}
		message.Text = "set priority 5"

		run, match := matcher.Match(message)
		assert.NotNil(t, run)
		assert.NotNil(t, match)

		if run != nil {
			run(match, message)
		}
		assert.True(t, executed)
	})

	t.Run("boolean parsing", func(t *testing.T) {
		executed := false
		matcher := NewRegexpMatcher(`set\s+debug\s+(?P<enabled>true|false)`, func(match Result, message msg.Message) {
			executed = true
			assert.True(t, match.Has("enabled"))
			assert.Equal(t, "true", match.GetString("enabled"))
		})

		message := msg.Message{}
		message.Text = "set debug true"

		run, match := matcher.Match(message)
		assert.NotNil(t, run)
		assert.NotNil(t, match)

		if run != nil {
			run(match, message)
		}
		assert.True(t, executed)
	})
}
