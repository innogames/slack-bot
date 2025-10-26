package matcher

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestGroupMatcherAdvanced(t *testing.T) {
	t.Run("single matcher optimization", func(t *testing.T) {
		// When there's only one matcher, it should return it directly
		singleMatcher := NewRegexpMatcher("test", testRunner)
		groupMatcher := NewGroupMatcher(singleMatcher)

		// The group matcher should be the same instance as the single matcher
		_, ok := groupMatcher.(regexpMatcher)
		assert.True(t, ok, "Single matcher should be returned directly")
	})

	t.Run("multiple matchers first match wins", func(t *testing.T) {
		executed := false
		matcher := NewGroupMatcher(
			NewRegexpMatcher("first.*", func(_ Result, message msg.Message) {
				executed = true
			}),
			NewRegexpMatcher("second pattern", func(_ Result, message msg.Message) {
				t.Error("Second matcher should not execute")
			}),
		)

		message := msg.Message{}
		message.Text = "first pattern with extra"

		run, match := matcher.Match(message)
		assert.NotNil(t, run)
		assert.NotNil(t, match)

		if run != nil {
			run(match, message)
		}
		assert.True(t, executed)
	})

	t.Run("no match in group", func(t *testing.T) {
		matcher := NewGroupMatcher(
			NewRegexpMatcher("pattern1", testRunner),
			NewRegexpMatcher("pattern2", testRunner),
		)

		message := msg.Message{}
		message.Text = "different pattern"

		run, match := matcher.Match(message)
		assert.Nil(t, run)
		assert.Nil(t, match)
	})

	t.Run("empty group matcher", func(t *testing.T) {
		matcher := NewGroupMatcher()

		message := msg.Message{}
		message.Text = "any text"

		run, match := matcher.Match(message)
		assert.Nil(t, run)
		assert.Nil(t, match)
	})
}

func TestMatcherIntegration(t *testing.T) {
	t.Run("prefix with admin matcher", func(t *testing.T) {
		executed := false

		cfg := config.Config{}
		cfg.AdminUsers = []string{"UADMIN"}

		slackClient := mocks.NewSlackClient(t)

		// Set up AllUsers map for user resolution
		client.AllUsers = map[string]string{
			"UADMIN": "admin",
		}

		adminMatcher := NewAdminMatcher(
			cfg.AdminUsers,
			slackClient,
			NewPrefixMatcher("admin command", func(_ Result, message msg.Message) {
				executed = true
				// Just verify that the admin command executes - don't worry about exact match content
			}),
		)

		message := msg.Message{}
		message.User = "UADMIN"
		message.Text = "admin command extra params"

		run, match := adminMatcher.Match(message)
		assert.NotNil(t, run)
		assert.NotNil(t, match)

		// Execute without triggering reaction (admin should execute normally)
		if run != nil {
			run(match, message)
		}
		assert.True(t, executed)

		// Verify no unexpected calls were made
		slackClient.AssertExpectations(t)
	})

	t.Run("non-admin blocked from admin command", func(t *testing.T) {
		cfg := config.Config{}
		cfg.AdminUsers = []string{"UADMIN"}

		slackClient := mocks.NewSlackClient(t)

		// Set up AllUsers map - UREGULAR is not in admin list and won't be found
		client.AllUsers = map[string]string{
			"UADMIN": "admin",
			// UREGULAR not in AllUsers map, so GetUserIDAndName will return empty
		}

		adminMatcher := NewAdminMatcher(
			cfg.AdminUsers,
			slackClient,
			NewPrefixMatcher("admin command", testRunner),
		)

		message := msg.Message{}
		message.User = "UREGULAR"
		message.Text = "admin command test"

		run, match := adminMatcher.Match(message)
		// Should return a runner (the error handler) and empty result
		assert.NotNil(t, run)
		assert.Equal(t, Result{}, match)

		// Mock the error handler calls
		slackClient.On("AddReaction", mock.AnythingOfType("util.Reaction"), mock.AnythingOfType("msg.Message")).Return(nil)
		slackClient.On("ReplyError", message, mock.MatchedBy(func(err error) bool {
			return err.Error() == "sorry, you are no admin and not allowed to execute this command"
		})).Return("")

		// Execute the error handler
		if run != nil {
			run(match, message)
		}

		slackClient.AssertExpectations(t)
	})
}
