package matcher

import (
	"github.com/slack-go/slack"
)

// NewVoidMatcher  just do nothing (might be useful when a command is not loadable because dependencies are not configures)
func NewVoidMatcher() Matcher {
	return voidMatcher{}
}

type voidMatcher struct {
}

func (m voidMatcher) Match(event slack.MessageEvent) (Runner, Result) {
	var match MapResult

	return nil, match
}
