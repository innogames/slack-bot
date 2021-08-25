package tester

import (
	"testing"

	"github.com/gookit/color"

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
			"Hallo :smile:: how are you? :",
			"Hallo 😄: how are you? :",
		},
		{
			"<@here> is some `code`",
			"\x1b[1m@here\x1b[0m is some \x1b[100mcode\x1b[0m",
		},
		{
			"Click <https://example.com|here>",
			"Click \x1b[34m\x1b]8;;https://example.com\ahere\x1b]8;;\a\x1b[0m",
		},
		{
			"Click <https://example.com|here> or <https://example.com|here>",
			"Click \x1b[34m\x1b]8;;https://example.com\ahere\x1b]8;;\a\x1b[0m or \x1b[34m\x1b]8;;https://example.com\ahere\x1b]8;;\a\x1b[0m",
		},
		{
			"*Hallo* :smile:",
			"\x1b[1mHallo\x1b[0m 😄",
		},
		{
			"_italic_ _ no italic _",
			"\x1b[3mitalic\x1b[0m _ no italic _",
		},
	}
	for _, tt := range tests {
		actual := formatSlackMessage(tt.input)
		assert.Equal(t, tt.expected, actual, "input: "+tt.input)
	}
}
