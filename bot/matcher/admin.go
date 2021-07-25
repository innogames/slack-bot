package matcher

import (
	"errors"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/client"
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

func (m adminMatcher) Match(message msg.Message) (Runner, Result) {
	run, result := m.matcher.Match(message)
	if run == nil {
		// the wrapped command didn't match...ignore
		return nil, nil
	}

	if m.admins.Contains(message.User) {
		// valid admin -> execute the wrapped command
		return run, result
	}

	return func(match Result, message msg.Message) {
		m.slackClient.AddReaction("❌", message)
		m.slackClient.ReplyError(
			message,
			errors.New("sorry, you are no admin and not allowed to execute this command"),
		)
	}, Result{}
}
