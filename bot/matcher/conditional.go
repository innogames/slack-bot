package matcher

import (
	"github.com/innogames/slack-bot/bot/util"
	"github.com/slack-go/slack"
)

// todo(matze): rename to wildcardMatcher etc?
func NewConditionalMatcher(run conditionalRunner) Matcher {
	return conditionalMatcher{run}
}

type conditionalRunner func(event slack.MessageEvent) bool

type conditionalMatcher struct {
	run conditionalRunner
}

func (m conditionalMatcher) Match(event slack.MessageEvent) (Runner, Result) {
	var match MapResult

	if m.run(event) {
		match = make(MapResult, 1)
		match[util.FullMatch] = event.Text
	}

	return nil, match
}
