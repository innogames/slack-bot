package openai

import (
	"regexp"
)

// https://platform.openai.com/docs/models/
var maxTokens = map[string]int{
	"gpt-4-turbo": 128000,
	"gpt-4o":      128000,
	"gpt-4.1":     1047576,
	"gpt-5":       400000,
	"gpt-5.1":     400000,
	"gpt-5.2":     400000,
	"gpt-5.4":     1000000,
	"gpt-5.5":     1000000,
	"dummy-test":  100, // just for testing
}

var modelDateRe = regexp.MustCompile(`-\d{4}`)

// truncateMessages will truncate the messages to fit into the max tokens limit of the model
// we always try to keep the last message, so we will truncate the first (oldest) messages
func truncateMessages(model string, inputMessages []ChatMessage) ([]ChatMessage, int, int) {
	maxTokens := getMaxTokensForModel(model)
	currentTokens := 0
	truncatedMessages := 0

	// Walk newest→oldest so we always keep the most recent messages (including the current prompt).
	kept := make([]ChatMessage, 0, len(inputMessages))
	for i := len(inputMessages) - 1; i >= 0; i-- {
		message := inputMessages[i]
		tokens := estimateTokensForMessage(message.Content)
		if currentTokens+tokens >= maxTokens {
			truncatedMessages++
			continue
		}
		currentTokens += tokens
		kept = append(kept, message)
	}

	// Reverse kept to restore chronological order.
	for i, j := 0, len(kept)-1; i < j; i, j = i+1, j-1 {
		kept[i], kept[j] = kept[j], kept[i]
	}

	return kept, currentTokens, truncatedMessages
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
