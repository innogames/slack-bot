package matcher

import (
	"github.com/innogames/slack-bot/bot/util"
	"github.com/slack-go/slack"
	"strings"
)

// NewTextMatcher match messages by full text (case insensitive)
func NewTextMatcher(text string, run Runner) Matcher {
	return textMatcher{
		loweredText: strings.ToLower(text),
		run:         run,
	}
}

type textMatcher struct {
	loweredText string
	run         Runner
}

func (m textMatcher) Match(event slack.MessageEvent) (Runner, Result) {
	var match MapResult

	if strings.EqualFold(event.Text, m.loweredText) {
		match := MapResult{
			util.FullMatch: event.Text,
		}
		return m.run, match
	}

	return nil, match
}
