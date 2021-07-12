package util

import (
	"fmt"
	"github.com/hackebrot/turtle"
	"strings"
)

// Reaction representation for a reaction/emoji
type Reaction string

// ToSlackReaction trims potential ":"
func (r Reaction) ToSlackReaction() string {
	return strings.Trim(string(r), ":")
}

// FullName with surrounding ":"
func (r Reaction) FullName() string {
	return fmt.Sprintf(":%s:", r.ToSlackReaction())
}

// GetChar get the real string/unicode representation of the current reaction/emoji
func (r Reaction) GetChar() string {
	name := strings.Trim(string(r), ":")
	emoji, ok := turtle.Emojis[name]
	if !ok {
		return "?"
	}
	return emoji.Char
}

// UnicodeToReaction transform a unicode char, like "😄" into a Reaction type
func UnicodeToReaction(inputChar string) Reaction {
	emoji, ok := turtle.EmojisByChar[inputChar]
	if !ok {
		return "?"
	}

	return Reaction(emoji.Name)
}
