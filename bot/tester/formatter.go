package tester

import (
	"regexp"
	"strings"

	"github.com/gookit/color"
	"github.com/hackebrot/turtle"
)

var boldRe = regexp.MustCompile(`\*.+\*`)
var emojiRe = regexp.MustCompile(`:.+:`)

func formatSlackMessage(msg string) string {
	msg = boldRe.ReplaceAllStringFunc(msg, func(part string) string {
		return color.Bold.Sprintf(strings.Trim(part, "*"))
	})

	msg = emojiRe.ReplaceAllStringFunc(msg, getEmoji)

	return msg
}

func getEmoji(name string) string {
	name = strings.Trim(name, ":")
	emoji, ok := turtle.Emojis[name]
	if !ok {
		return "??"
	}
	return emoji.Char
}
