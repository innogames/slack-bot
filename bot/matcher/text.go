package matcher

import (
	"strings"

	"github.com/innogames/slack-bot/v2/bot/msg"
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
		return m.run, Result{}
	}

	return nil, nil
}
