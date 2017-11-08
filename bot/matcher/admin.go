package matcher

import (
	"fmt"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/config"
	"github.com/nlopes/slack"
)

// NewAdminMatcher is a wrapper to only executable by a whitelisted admin user
func NewAdminMatcher(cfg config.Config, slackClient client.SlackClient, matcher Matcher) Matcher {
	return adminMatcher{matcher, cfg, slackClient}
}

type adminMatcher struct {
	matcher     Matcher
	cfg         config.Config
	slackClient client.SlackClient
}

func (m adminMatcher) Match(event slack.MessageEvent) (Runner, Result) {
	run, result := m.matcher.Match(event)
	if !result.Matched() {
		return nil, result
	}

	for _, adminId := range m.cfg.AdminUsers {
		if adminId == event.User {
			return run, result
		}
	}

	match := MapResult{
		util.FullMatch: event.Text,
	}

	// todo use logger
	fmt.Printf("Command was executed without admin access: %s - %s\n", event.Text, event.User)
	return func(match Result, event slack.MessageEvent) {
		m.slackClient.Reply(event, "Sorry, you are no admin and not allowed to execute this command!")
	}, match
}
