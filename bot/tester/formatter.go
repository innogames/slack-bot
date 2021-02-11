package tester

import (
	"github.com/innogames/slack-bot/bot/util"
	"regexp"
	"strings"

	"github.com/gookit/color"
)

var boldRe = regexp.MustCompile(`\*.+\*`)
var emojiRe = regexp.MustCompile(`:.+:`)

func formatSlackMessage(msg string) string {
	msg = boldRe.ReplaceAllStringFunc(msg, func(part string) string {
		return color.Bold.Sprintf(strings.Trim(part, "*"))
	})

	msg = emojiRe.ReplaceAllStringFunc(msg, func(s string) string {
		return util.Reaction(s).GetChar()
	})

	return msg
}
