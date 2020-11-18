package matcher

import (
	"errors"

	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/innogames/slack-bot/client"
	"github.com/slack-go/slack"
)

// NewAdminMatcher is a wrapper to only executable by a whitelisted admins user
func NewAdminMatcher(admins config.UserList, slackClient client.SlackClient, matcher Matcher) Matcher {
	return adminMatcher{matcher, admins, slackClient}
}

type adminMatcher struct {
	matcher     Matcher
	admins      config.UserList
	slackClient client.SlackClient
}

func (m adminMatcher) Match(event slack.MessageEvent) (Runner, Result) {
	run, result := m.matcher.Match(event)
	if !result.Matched() {
		return nil, result
	}

	if m.admins.Contains(event.User) {
		// valid admin -> execute the wrapped command
		return run, result
	}

	match := MapResult{
		util.FullMatch: event.Text,
	}

	return func(match Result, event slack.MessageEvent) {
		m.slackClient.ReplyError(
			event,
			errors.New("sorry, you are no admins and not allowed to execute this command"),
		)
	}, match
}
