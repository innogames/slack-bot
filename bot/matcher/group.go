package matcher

import (
	"github.com/innogames/slack-bot/v2/bot/msg"
)

// NewGroupMatcher is a matcher that iterates through the list of specified sub-matchers ...just define multiple matcher in a chain/group
func NewGroupMatcher(matcher ...Matcher) Matcher {
	return groupMatcher{
		matcher: matcher,
	}
}

type groupMatcher struct {
	matcher []Matcher
}

func (m groupMatcher) Match(message msg.Message) (Runner, Result) {
	for _, matcher := range m.matcher {
		run, match := matcher.Match(message)
		if match != nil {
			return run, match
		}
	}

	return nil, nil
}
