package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReaction(t *testing.T) {
	smile := Reaction(":smile:")
	assert.Equal(t, "smile", smile.ToSlackReaction())
	assert.Equal(t, ":smile:", smile.FullName())
	assert.Equal(t, "😄", smile.GetChar())

	unknown := Reaction(":unknown:")
	assert.Equal(t, "unknown", unknown.ToSlackReaction())
	assert.Equal(t, "?", unknown.GetChar())
}

func TestEmojiToReaction(t *testing.T) {
	smile := UnicodeToReaction("😄")
	assert.Equal(t, "smile", smile.ToSlackReaction())
	assert.Equal(t, ":smile:", smile.FullName())
	assert.Equal(t, "😄", smile.GetChar())

	unknown := UnicodeToReaction("empty")
	assert.Equal(t, "?", unknown.ToSlackReaction())
	assert.Equal(t, "?", unknown.GetChar())
}
