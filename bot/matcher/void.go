package matcher

import (
	"github.com/innogames/slack-bot/v2/bot/msg"
)

// NewVoidMatcher just do nothing (might be useful when a command is not loadable because dependencies are not configures)
func NewVoidMatcher() Matcher {
	return voidMatcher{}
}

type voidMatcher struct{}

func (m voidMatcher) Match(message msg.Message) (Runner, Result) {
	return nil, nil
}
