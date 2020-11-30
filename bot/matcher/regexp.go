package matcher

import (
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/util"
	"regexp"
)

// NewRegexpMatcher matches a command based on a given prefix. 2 additional rules:
// - it's case insensitive
// - it always has to match the full line (adding ^ and $ implicitly)
func NewRegexpMatcher(regexpString string, run Runner) Matcher {
	return &regexpMatcher{
		regexp: util.CompileRegexp(regexpString),
		run:    run,
	}
}

type regexpMatcher struct {
	regexp *regexp.Regexp
	run    Runner
}

func (m *regexpMatcher) Match(message msg.Message) (Runner, Result) {
	var match MapResult

	matches := m.regexp.FindStringSubmatch(message.GetText())
	if len(matches) == 0 {
		return nil, match
	}

	return m.run, ReResult{matches, m.regexp}
}
