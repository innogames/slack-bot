package tester

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEmoji(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			"smile",
			"😄",
		},
		{
			":smile:",
			"😄",
		},
		{
			"Idontknpw",
			"??",
		},
		{
			"",
			"??",
		},
	}
	for _, tt := range tests {
		actual := getEmoji(tt.name)
		assert.Equal(t, tt.want, actual, "input: "+tt.name)
	}
}

func TestPrintSlackMessage(t *testing.T) {
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
	}
	for _, tt := range tests {
		actual := formatSlackMessage(tt.input)
		assert.Equal(t, tt.expected, actual, "input: "+tt.input)
	}
}
