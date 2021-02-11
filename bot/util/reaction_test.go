package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReaction(t *testing.T) {
	smile := Reaction(":smile:")
	assert.Equal(t, "smile", smile.ToSlackReaction())
	assert.Equal(t, "😄", smile.GetChar())

	unknown := Reaction(":unknown:")
	assert.Equal(t, "unknown", unknown.ToSlackReaction())
	assert.Equal(t, "?", unknown.GetChar())
}
