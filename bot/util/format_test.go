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
	t.Run("FormatBytes", func(t *testing.T) {
		for _, testCase := range formatTestCases {
			actual := FormatBytes(testCase.input)
			assert.Equal(t, testCase.expected, actual)
		}
	})
}

var formatIntTestCases = []struct {
	input    int
	expected string
}{
	{-1000, "-1,000"},
	{0, "0"},
	{1, "1"},
	{100, "100"},
	{999, "999"},
	{1000, "1,000"},
	{232632, "232,632"},
	{3627838232862387286, "3,627,838,232,862,387,286"},
}

func TestFormatInt(t *testing.T) {
	t.Run("TestFormatInt", func(t *testing.T) {
		for _, testCase := range formatIntTestCases {
			actual := FormatInt(testCase.input)
			assert.Equal(t, testCase.expected, actual)
		}
	})
}
