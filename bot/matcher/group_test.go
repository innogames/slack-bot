package matcher

import (
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/mocks"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGroup(t *testing.T) {
	event := slack.MessageEvent{}
	event.User = "UADMIN"

	cfg := config.Config{}
	cfg.AdminUsers = []string{
		"UADMIN",
	}

	matcher := NewGroupMatcher(
		NewRegexpMatcher(`add reaction :(?P<reaction>.*):`, testRunner),
		NewAdminMatcher(
			cfg.AdminUsers,
			&mocks.SlackClient{},
			NewRegexpMatcher(`remove reaction :(?P<reaction>.*):`, testRunner),
		),
	)

	t.Run("Match simple", func(t *testing.T) {
		var matchTest = []struct {
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
			event.Text = testCase.input
			run, result := matcher.Match(event)
			if testCase.expected {
				assert.NotNil(t, run)
				assert.True(t, result.Matched())
			} else {
				assert.Nil(t, run)
				assert.False(t, result.Matched())
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
			&mocks.SlackClient{},
			NewTextMatcher(`text`, testRunner),
		),
		NewVoidMatcher(),
	)
	b.Run("chained prefix: no match", func(b *testing.B) {
		event := slack.MessageEvent{}
		event.Text = "haha reaction :foo:"

		for i := 0; i < b.N; i++ {
			regexpChainMatcher.Match(event)
		}
	})

	b.Run("chained prefix: match", func(b *testing.B) {
		event := slack.MessageEvent{}
		event.Text = "add reaction :foo:"

		for i := 0; i < b.N; i++ {
			regexpChainMatcher.Match(event)
		}
	})
}
