package tester

import (
	"github.com/gookit/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrintSlackMessage(t *testing.T) {
	color.ForceColor()

	tests := []struct {
		input    string
		expected string
	}{
		{
			"",
			"",
		},
		{
			"Hallo :smile:",
			"Hallo 😄",
		},
		{
			"*Hallo* :smile:",
			"\x1b[1mHallo\x1b[0m 😄",
		},
	}
	for _, tt := range tests {
		actual := formatSlackMessage(tt.input)
		assert.Equal(t, tt.expected, actual, "input: "+tt.input)
	}
}
