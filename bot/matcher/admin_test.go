package matcher

import (
	"errors"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/mocks"
	"github.com/slack-go/slack"
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

		testRunner := func(match Result, event slack.MessageEvent) {}
		matcher := NewTextMatcher("test", testRunner)
		subject := NewAdminMatcher(cfg, slackClient, matcher)

		event := slack.MessageEvent{}
		event.Text = "test"

		slackClient.On("ReplyError", event, errors.New("sorry, you are no admin and not allowed to execute this command"))

		runner, match := subject.Match(event)
		runner(MapResult{}, event)

		assert.True(t, match.Matched())
	})
}
