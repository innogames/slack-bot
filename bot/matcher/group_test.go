package matcher

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
)

func TestGroup(t *testing.T) {
	message := msg.Message{}
	message.User = "UADMIN"

	cfg := config.Config{}
	cfg.AdminUsers = []string{
		"UADMIN",
	}

	t.Run("Match empty", func(t *testing.T) {
		matcher := NewGroupMatcher()

		run, result := matcher.Match(message)
		assert.Nil(t, run)
		assert.Nil(t, result)
	})
	t.Run("Match One", func(t *testing.T) {
		matcher := NewGroupMatcher(
			NewVoidMatcher(),
		)

		run, result := matcher.Match(message)
		assert.Nil(t, run)
		assert.Nil(t, result)
	})

	t.Run("Match multi cacses", func(t *testing.T) {
		matcher := NewGroupMatcher(
			NewRegexpMatcher(`add reaction :(?P<reaction>.*):`, testRunner),
			NewAdminMatcher(
				cfg.AdminUsers,
				mocks.NewSlackClient(t),
				NewRegexpMatcher(`remove reaction :(?P<reaction>.*):`, testRunner),
			),
		)

		matchTest := []struct {
			input    string
			expected bool
		}{
			{"", false},
			{"sdsd", false},
			{"add reaction :foo", false},
			{"add reaction :foo: ", false},
			{"add reaction :foo:", true},
			{"add reaction :foo:", true},
			{"remove reaction :foo:", true},
		}

		for _, testCase := range matchTest {
			message.Text = testCase.input
			run, result := matcher.Match(message)
			if testCase.expected {
				assert.NotNil(t, run)
				assert.NotNil(t, result)
			} else {
				assert.Nil(t, run)
				assert.Nil(t, result)
			}
		}
	})
}

func BenchmarkMatchChained(b *testing.B) {
	regexpChainMatcher := NewGroupMatcher(
		NewRegexpMatcher(`add reaction :(?P<reaction>.*):`, testRunner),
		NewRegexpMatcher(`remove reaction :(?P<reaction>.*):`, testRunner),
		NewPrefixMatcher(`prefix`, testRunner),
		NewAdminMatcher(
			config.UserList{},
			mocks.NewSlackClient(b),
			NewTextMatcher(`text`, testRunner),
		),
		NewVoidMatcher(),
	)
	b.Run("chained prefix: no match", func(b *testing.B) {
		message := msg.Message{}
		message.Text = "haha reaction :foo:"

		for range b.N {
			regexpChainMatcher.Match(message)
		}
	})

	b.Run("chained prefix: match", func(b *testing.B) {
		message := msg.Message{}
		message.Text = "add reaction :foo:"

		for range b.N {
			regexpChainMatcher.Match(message)
		}
	})
}
