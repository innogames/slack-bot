package matcher

import (
	"github.com/innogames/slack-bot/bot/util"
	"github.com/nlopes/slack"
	"strings"
)

// NewPrefixMatcher accepts command which starts with the given prefix
// Example: prefix for "random"
// "random" -> match
// "random 1 2 3" -> match
// "randomness" -> no match
func NewPrefixMatcher(prefix string, run Runner) Matcher {
	return prefixMatcher{
		strings.ToLower(prefix),
		run,
	}
}

type prefixMatcher struct {
	loweredPrefix string
	run           Runner
}

func (m prefixMatcher) Match(event slack.MessageEvent) (Runner, Result) {
	var match MapResult

	if strings.HasPrefix(strings.ToLower(event.Text), m.loweredPrefix) {
		suffix := event.Text[len(m.loweredPrefix):]
		// avoid matching "randomness" if prefix is "random"
		if len(suffix) > 0 && suffix[0] != ' ' {
			return nil, match
		}

		match := MapResult{
			util.FullMatch: strings.TrimPrefix(suffix, " "),
		}
		return m.run, match
	}

	return nil, match
}
