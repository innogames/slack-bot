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
		{"gpt-5.5", 1000000},
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
	assert.Equal(t, 99, inputTokens) // newest 5 messages fit; msg[0] and msg[2] are dropped
	assert.Equal(t, 2, truncatedMessages)

	// The newest (last) messages must survive truncation — we keep the tail, not the head.
	assert.Equal(t, messages[len(messages)-1], outputMessages[len(outputMessages)-1], "last message must be kept")
	assert.Equal(t, messages[len(messages)-2], outputMessages[len(outputMessages)-2], "second-to-last message must be kept")
}

func TestTruncateKeepsNewest(t *testing.T) {
	// dummy-test model has max 100 tokens, each char≈0.25 tokens (estimateTokensForMessage divides by 4)
	// "system" = 6 chars = 1 token, "current user question" = 21 chars = 5 tokens
	// Pack old messages to fill the budget, then verify the newest (current prompt) is still present.
	old := ChatMessage{Content: "old history old history old history old history old history old history old history old history old"} // 99 chars ≈ 24 tokens * 4 = 96
	current := ChatMessage{Content: "current user question"}

	inputMessages := []ChatMessage{old, old, old, old, current}

	outputMessages, _, truncated := truncateMessages("dummy-test", inputMessages)

	assert.Positive(t, truncated, "some old messages should be truncated")
	assert.Equal(t, current, outputMessages[len(outputMessages)-1], "current (newest) message must survive truncation")
}

func TestCountTokens(t *testing.T) {
	t.Run("Count", func(t *testing.T) {
		actual := estimateTokensForMessage("hello you!")
		assert.Equal(t, 2, actual)
	})
}
