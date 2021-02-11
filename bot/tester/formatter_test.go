package tester

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
