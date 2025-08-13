package openai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModels(t *testing.T) {
	modelsTestCases := []struct {
		input    string
		expected int
	}{
		{"", 128000},
		{"jolo", 128000},
		{"gpt-4.1", 1047576},
		{"gpt-4o", 128000},
		{"gpt-5-mini", 128000},
		{"gpt-3.5-turbo", 16385},
	}

	for _, testCase := range modelsTestCases {
		actual := getMaxTokensForModel(testCase.input)
		assert.Equal(t, testCase.expected, actual, "Model "+testCase.input)
	}
}

func TestTruncate(t *testing.T) {
	messages := []ChatMessage{
		{Content: "hello, Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."},
		{Content: "hello i am a super long string, with many tokens"},
		{Content: "hello i am a super long string with many tokens, foo bar baz"},
		{Content: "hello i am a super long string with many tokens, foo bar baz"},
		{Content: "or am i? do you think so? and what about this?"},
		{Content: "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."},
		{Content: "Lorem ipsum domino sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."},
	}

	outputMessages, inputTokens, truncatedMessages := truncateMessages("dummy-test", messages)

	assert.Len(t, outputMessages, 5)
	assert.Equal(t, 85, inputTokens)
	assert.Equal(t, 2, truncatedMessages)
}

func TestCountTokens(t *testing.T) {
	t.Run("Count", func(t *testing.T) {
		actual := estimateTokensForMessage("hello you!")
		assert.Equal(t, 2, actual)
	})
}
