package openai

import (
	"regexp"

	"github.com/tiktoken-go/tokenizer"
)

// https://platform.openai.com/docs/models/gpt-3-5
// https://platform.openai.com/docs/models/gpt-4
var maxTokens = map[string]int{
	"gpt-4":             8192,
	"gpt-4-32k":         32768,
	"gpt-3.5-turbo-16k": 16385,
	"gpt-3.5-turbo":     4097,
	"dummy-test":        100, // just for testing
}

var modelDateRe = regexp.MustCompile("-\\d{4}")

func truncateMessages(model string, inputMessages []ChatMessage) ([]ChatMessage, int, int) {
	outputMessages := make([]ChatMessage, 0, len(inputMessages))

	currentTokens := 0
	truncatedMessages := 0
	maxTokens := getMaxTokensForModel(model)
	for _, message := range inputMessages {
		tokens := countTokensForMessage(message)

		if currentTokens+tokens >= maxTokens {
			truncatedMessages++
			continue
		}
		currentTokens += tokens
		outputMessages = append(outputMessages, message)
	}

	return outputMessages, currentTokens, truncatedMessages
}

func getMaxTokensForModel(model string) int {
	if maxTokens, ok := maxTokens[model]; ok {
		return maxTokens
	}

	if modelDateRe.MatchString(model) {
		return getMaxTokensForModel(modelDateRe.ReplaceAllString(model, ""))
	}

	// we need some default
	return 4000
}

func countTokensForMessage(message ChatMessage) int {
	enc, _ := tokenizer.ForModel(tokenizer.GPT4)
	tokens, _, _ := enc.Encode(message.Content)

	return len(tokens)
}
