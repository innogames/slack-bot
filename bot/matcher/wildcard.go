package matcher

import (
	"github.com/innogames/slack-bot/bot/util"
	"github.com/slack-go/slack"
)

// WildcardMatcher will pass all messages into the runner. The runner needs to check if the event is relevant or not
// if the runner returns "true", the command is marked as executed and no other commands are checked
func WildcardMatcher(run wildcardRunner) Matcher {
	return wildcardMatcher{run}
}

type wildcardRunner func(event slack.MessageEvent) bool

type wildcardMatcher struct {
	run wildcardRunner
}

func (m wildcardMatcher) Match(event slack.MessageEvent) (Runner, Result) {
	var match MapResult

	if m.run(event) {
		match = make(MapResult, 1)
		match[util.FullMatch] = event.Text
	}

	return nil, match
}
