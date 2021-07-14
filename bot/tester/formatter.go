package tester

import (
	"regexp"
	"strings"

	"github.com/innogames/slack-bot.v2/bot/util"

	"github.com/gookit/color"
)

var (
	boldRe  = regexp.MustCompile(`\*.+\*`)
	emojiRe = regexp.MustCompile(`:.+:`)
)

func formatSlackMessage(msg string) string {
	msg = boldRe.ReplaceAllStringFunc(msg, func(part string) string {
		return color.Bold.Sprintf(strings.Trim(part, "*"))
	})

	msg = emojiRe.ReplaceAllStringFunc(msg, func(s string) string {
		return util.Reaction(s).GetChar()
	})

	return msg
}
