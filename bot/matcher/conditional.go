package matcher

import (
	"github.com/innogames/slack-bot/bot/util"
	"github.com/slack-go/slack"
)

func NewConditionalMatcher(run conditionalRunner) Matcher {
	return conditionalMather{run}
}

type conditionalRunner func(event slack.MessageEvent) bool

type conditionalMather struct {
	run conditionalRunner
}

func (m conditionalMather) Match(event slack.MessageEvent) (Runner, Result) {
	var match MapResult

	if m.run(event) {
		match = make(MapResult, 1)
		match[util.FullMatch] = event.Text
	}

	return nil, match
}
