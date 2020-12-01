package matcher

import (
	"github.com/innogames/slack-bot/bot/msg"
)

// NewGroupMatcher is a matcher which go through the list of given sub-matcher...just define multiple matcher in a chain/group
func NewGroupMatcher(matcher ...Matcher) Matcher {
	return groupMatcher{
		matcher: matcher,
	}
}

type groupMatcher struct {
	matcher []Matcher
}

func (m groupMatcher) Match(message msg.Message) (Runner, Result) {
	var match MapResult

	for _, matcher := range m.matcher {
		run, match := matcher.Match(message)
		if match.Matched() {
			return run, match
		}
	}

	return nil, match
}
