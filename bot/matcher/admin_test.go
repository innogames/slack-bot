package matcher

import (
	"errors"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
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

		slackClient.On("ReplyError", message, errors.New("sorry, you are no admins and not allowed to execute this command"))

		runner, match := subject.Match(message)
		runner(MapResult{}, message)

		assert.True(t, match.Matched())
	})
}
