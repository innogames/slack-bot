package matcher

import (
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/util"
)

// WildcardMatcher will pass all messages into the runner. The runner needs to check if the event is relevant or not
// if the runner returns "true", the command is marked as executed and no other commands are checked
func WildcardMatcher(run wildcardRunner) Matcher {
	return wildcardMatcher{run}
}

type wildcardRunner func(ref msg.Ref, text string) bool

type wildcardMatcher struct {
	run wildcardRunner
}

func (m wildcardMatcher) Match(message msg.Message) (Runner, Result) {
	var match MapResult

	if m.run(message, message.GetText()) {
		match = make(MapResult, 1)
		match[util.FullMatch] = message.GetText()
	}

	return nil, match
}
