package openai

import (
	"regexp"
)

// https://platform.openai.com/docs/models/gpt-3-5
// https://platform.openai.com/docs/models/gpt-4
var maxTokens = map[string]int{
	"gpt-4":                8192,
	"gpt-4-32k":            32768,
	"gpt-3.5-turbo-16k":    16385,
	"gpt-3.5-turbo":        4096,
	"gpt-4-1106-preview":   128000,
	"gpt-4-vision-preview": 128000,
	"dummy-test":           100, // just for testing
}

var modelDateRe = regexp.MustCompile(`-\d{4}`)

func truncateMessages(model string, inputMessages []ChatMessage) ([]ChatMessage, int, int) {
	outputMessages := make([]ChatMessage, 0, len(inputMessages))

	currentTokens := 0
	truncatedMessages := 0
	maxTokens := getMaxTokensForModel(model)
	for _, message := range inputMessages {
		tokens := estimateTokensForMessage(message.Content)

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

func estimateTokensForMessage(message string) int {
	// to lower the dependency to heavy external libs we use the rule of thumbs which is totally fine here
	// https://platform.openai.com/tokenizer
	return len(message) / 4
}
