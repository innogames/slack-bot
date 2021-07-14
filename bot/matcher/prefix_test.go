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
