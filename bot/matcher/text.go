package matcher

import (
	"strings"

	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/util"
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

func (m textMatcher) Match(message msg.Message) (Runner, Result) {
	if strings.EqualFold(message.Text, m.loweredText) {
		match := Result{
			util.FullMatch: message.Text,
		}
		return m.run, match
	}

	return nil, nil
}
