package matcher

import (
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPrefix(t *testing.T) {
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
				assert.Equal(t, testCase.match, match.MatchedString())
			} else {
				assert.Nil(t, run, testCase.input)
			}
			assert.Equal(t, testCase.expected, match.Matched())
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
