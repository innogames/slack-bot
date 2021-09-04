package tester

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gookit/color"
	"github.com/innogames/slack-bot/v2/bot/util"
)

var (
	boldRe   = regexp.MustCompile(`\*[^\s][^\*]*[^\s]\*`)
	italicRe = regexp.MustCompile(`_[^\s][^_]*[^\s]_`)
	linkRe   = regexp.MustCompile(`<(.+?)\|([^>]+)>`)
	codeRe   = regexp.MustCompile("`+([^`]+)`+")
	emojiRe  = regexp.MustCompile(`:[\w_]+:`)
	userRe   = regexp.MustCompile(`<@[\w_]+>`)
)

// supports:
// - bold/italic
// - emojis
// - `code` blocks
// - links (as blue colored ANSI hyperlink)
func formatSlackMessage(msg string) string {
	if color.Enable {
		msg = linkRe.ReplaceAllString(msg, color.Blue.Sprint("\u001B]8;;$1\u0007$2\u001B]8;;\u0007"))
	}

	msg = emojiRe.ReplaceAllStringFunc(msg, func(s string) string {
		return util.Reaction(s).GetChar()
	})

	msg = userRe.ReplaceAllStringFunc(msg, func(part string) string {
		return color.Bold.Sprintf(strings.Trim(part, "<>"))
	})

	msg = boldRe.ReplaceAllStringFunc(msg, func(part string) string {
		return color.Bold.Sprintf(strings.Trim(part, "*"))
	})

	msg = codeRe.ReplaceAllStringFunc(msg, func(part string) string {
		return color.BgGray.Sprintf(strings.Trim(part, "`"))
	})

	msg = italicRe.ReplaceAllStringFunc(msg, func(part string) string {
		return color.OpItalic.Sprintf(strings.Trim(part, "_"))
	})

	return msg
}

func commandButton(text string, command string) string {
	return fmt.Sprintf("<%scommand?command=%s|%s>", FakeServerURL, command, text)
}
