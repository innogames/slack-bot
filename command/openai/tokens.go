package openai

import (
	"regexp"
)

// https://platform.openai.com/docs/models/gpt-3-5
// https://platform.openai.com/docs/models/gpt-4
var maxTokens = map[string]int{
	"gpt-4":              8192,
	"gpt-4-32k":          32768,
	"gpt-4-1106-preview": 128000,
	"gpt-4-turbo":        128000,
	"gpt-4o":             128000,
	"gpt-3.5-turbo":      16385,
	"dummy-test":         100, // just for testing
}

var modelDateRe = regexp.MustCompile(`-\d{4}`)

// truncateMessages will truncate the messages to fit into the max tokens limit of the model
// we always try to keep the last message, so we will truncate the first messages
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

	// we need some default, keep it high, as new models will most likely support more tokens
	return 128000
}

// to lower the dependency to heavy external libs we use the rule of thumbs which is totally fine here
// https://platform.openai.com/tokenizer
func estimateTokensForMessage(message string) int {
	return len(message) / 4
}
