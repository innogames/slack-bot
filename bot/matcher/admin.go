package matcher

import (
	"errors"

	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/innogames/slack-bot/client"
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
	if !result.Matched() {
		// the wrapped command didn't match...ignore
		return nil, result
	}

	if m.admins.Contains(message.User) {
		// valid admin -> execute the wrapped command
		return run, result
	}

	match := MapResult{
		util.FullMatch: message.Text,
	}

	return func(match Result, message msg.Message) {
		m.slackClient.ReplyError(
			message,
			errors.New("sorry, you are no admin and not allowed to execute this command"),
		)
	}, match
}
