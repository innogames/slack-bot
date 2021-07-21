package util

import (
	"strings"

	"github.com/hackebrot/turtle"
)

// Reaction representation for a reaction/emoji. It can be the "id" of the reaction (like "smile", or the actual Unicode char, like "😄")
type Reaction string

// ToSlackReaction uses the "id" of the reaction/Emoji, like "smile". It trims potential ":"
func (r Reaction) ToSlackReaction() string {
	emoji := r.getEmoji()
	if emoji == nil {
		return strings.Trim(string(r), ":")
	}

	return emoji.Name
}

// GetChar get the real string/unicode representation of the current reaction/emoji
func (r Reaction) GetChar() string {
	emoji := r.getEmoji()
	if emoji == nil {
		return "?"
	}

	return emoji.Char
}

func (r Reaction) getEmoji() *turtle.Emoji {
	name := strings.Trim(string(r), ":")
	if emoji, ok := turtle.Emojis[name]; ok {
		return emoji
	}
	if emoji, ok := turtle.EmojisByChar[name]; ok {
		return emoji
	}

	return nil
}
