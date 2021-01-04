package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var formatTestCases = []struct {
	input    uint64
	expected string
}{
	{0, "0 B"},
	{1, "1 B"},
	{100, "100 B"},
	{999, "999 B"},
	{1000, "1.0 kB"},
	{232632, "232.6 kB"},
	{3627838232862387286, "3.6 EB"},
}

func TestFormatBytes(t *testing.T) {
	t.Run("ParseDuration", func(t *testing.T) {
		for _, testCase := range formatTestCases {
			actual := FormatBytes(testCase.input)
			assert.Equal(t, testCase.expected, actual)
		}
	})
}
