package matcher

import (
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/util"
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

func (m prefixMatcher) Match(message msg.Message) (Runner, Result) {
	var match MapResult

	if strings.HasPrefix(strings.ToLower(message.Text), m.loweredPrefix) {
		suffix := message.Text[len(m.loweredPrefix):]
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
