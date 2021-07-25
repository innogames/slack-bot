package matcher

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
)

func TestAdmin(t *testing.T) {
	t.Run("Test no Admin", func(t *testing.T) {
		slackClient := &mocks.SlackClient{}

		cfg := config.Config{}
		cfg.AdminUsers = []string{
			"UADMIN",
		}

		testRunner := func(match Result, message msg.Message) {}
		matcher := NewTextMatcher("test", testRunner)
		subject := NewAdminMatcher(cfg.AdminUsers, slackClient, matcher)

		message := msg.Message{}
		message.Text = "test"

		mocks.AssertReaction(slackClient, "❌", message)
		mocks.AssertError(slackClient, message, "sorry, you are no admin and not allowed to execute this command")

		runner, match := subject.Match(message)
		runner(Result{}, message)

		assert.NotNil(t, match)
		assert.NotNil(t, runner)
	})
}
