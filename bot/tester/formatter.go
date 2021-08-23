package tester

import (
	"regexp"
	"strings"

	"github.com/gookit/color"
	"github.com/innogames/slack-bot/v2/bot/util"
)

var (
	boldRe   = regexp.MustCompile(`\*[^\s][^\*]*[^\s]\*`)
	italicRe = regexp.MustCompile(`_[^\s][^_]*[^\s]_`)
	linkRe   = regexp.MustCompile(`<(.+)\|([^>]+)>`)
	emojiRe  = regexp.MustCompile(`:[\w_]+:`)
)

// supports:
// - bold/italic
// - emojis
// - links (as blue colored ANSI hyperlink)
func formatSlackMessage(msg string) string {
	if color.Enable {
		msg = linkRe.ReplaceAllString(msg, color.Blue.Sprint("\u001B]8;;$1\u0007$2\u001B]8;;\u0007"))
	}

	msg = boldRe.ReplaceAllStringFunc(msg, func(part string) string {
		return color.Bold.Sprintf(strings.Trim(part, "*"))
	})

	msg = italicRe.ReplaceAllStringFunc(msg, func(part string) string {
		return color.OpItalic.Sprintf(strings.Trim(part, "_"))
	})

	msg = emojiRe.ReplaceAllStringFunc(msg, func(s string) string {
		return util.Reaction(s).GetChar()
	})

	return msg
}
