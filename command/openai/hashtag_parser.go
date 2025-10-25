package openai

import (
	"regexp"
	"strconv"
	"strings"
)

// HashtagOptions contains parsed hashtag options from user input
type HashtagOptions struct {
	ReasoningEffort string // "minimal", "medium", "high", or ""
	Model           string // override model, empty means use config default
	MessageHistory  int    // number of channel messages to include, 0 means disabled
	NoStreaming     bool   // disable streaming responses, get full response at once
	NoThread        bool   // disable thread replies, reply directly to the message instead
}

// ParseHashtags extracts hashtag options from the input text and returns
// the cleaned text (without hashtags) and the parsed options
func ParseHashtags(text string) (cleanText string, options HashtagOptions) {
	// Define hashtag patterns
	// Pattern priorities (specific to general):
	// 1. #message-history-<number>
	// 2. #message-history
	// 3. #model-<name>
	// 4. #high-thinking, #medium-thinking, #minimal-thinking, #no-thinking

	// Parse message-history with number: #message-history-20
	messageHistoryWithCountRe := regexp.MustCompile(`#message-history-(\d+)`)
	if matches := messageHistoryWithCountRe.FindStringSubmatch(text); len(matches) > 1 {
		if count, err := strconv.Atoi(matches[1]); err == nil {
			options.MessageHistory = count
		}
		text = messageHistoryWithCountRe.ReplaceAllString(text, "")
	} else if strings.Contains(text, "#message-history") {
		// Parse message-history without number (default to 10)
		options.MessageHistory = 10
		text = strings.ReplaceAll(text, "#message-history", "")
	}

	// Parse model: #model-gpt-4o
	modelRe := regexp.MustCompile(`#model-([\w.-]+)`)
	if matches := modelRe.FindStringSubmatch(text); len(matches) > 1 {
		options.Model = matches[1]
		text = modelRe.ReplaceAllString(text, "")
	}

	// Parse reasoning effort
	switch {
	case strings.Contains(text, "#high-thinking"):
		options.ReasoningEffort = "high"
		text = strings.ReplaceAll(text, "#high-thinking", "")
	case strings.Contains(text, "#medium-thinking"):
		options.ReasoningEffort = "medium"
		text = strings.ReplaceAll(text, "#medium-thinking", "")
	case strings.Contains(text, "#minimal-thinking"):
		options.ReasoningEffort = "minimal"
		text = strings.ReplaceAll(text, "#minimal-thinking", "")
	case strings.Contains(text, "#no-thinking"):
		options.ReasoningEffort = "none"
		text = strings.ReplaceAll(text, "#no-thinking", "")
	}

	// Parse no-streaming option
	if strings.Contains(text, "#no-streaming") {
		options.NoStreaming = true
		text = strings.ReplaceAll(text, "#no-streaming", "")
	}

	// Parse no-thread option
	if strings.Contains(text, "#no-thread") {
		options.NoThread = true
		text = strings.ReplaceAll(text, "#no-thread", "")
	}

	// Clean up extra whitespace
	cleanText = strings.Join(strings.Fields(text), " ")

	return cleanText, options
}
