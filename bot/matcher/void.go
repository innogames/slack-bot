package matcher

import (
	"github.com/nlopes/slack"
)

// NewVoidMatcher  just do nothing (might be useful when a command is not loadable because dependencies are not configures)
func NewVoidMatcher() Matcher {
	return voidMatcher{}
}

type voidMatcher struct {
	run conditionalRunner
}

func (m voidMatcher) Match(event slack.MessageEvent) (Runner, Result) {
	var match MapResult

	return nil, match
}
