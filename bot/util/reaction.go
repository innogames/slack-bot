package util

import (
	"github.com/hackebrot/turtle"
	"strings"
)

// Reaction representation for a reaction/emoji
type Reaction string

// ToSlackReaction trims potential ":"
func (r Reaction) ToSlackReaction() string {
	return strings.Trim(string(r), ":")
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
