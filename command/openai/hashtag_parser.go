package openai

import (
	"regexp"
	"strconv"
	"strings"
)

// removeHashtag removes a hashtag from the text if present and returns whether it was found
func removeHashtag(text *string, hashtag string) bool {
	if strings.Contains(*text, hashtag) {
		*text = strings.ReplaceAll(*text, hashtag, "")
		return true
	}
	return false
}

// HashtagOptions contains parsed hashtag options from user input
type HashtagOptions struct {
	ReasoningEffort string // "minimal", "medium", "high", or ""
	Model           string // override model, empty means use config default
	MessageHistory  int    // number of channel messages to include, 0 means disabled
	NoStreaming     bool   // disable streaming responses, get full response at once
	NoThread        bool   // disable thread replies, reply directly to the message instead
	Debug           bool   // show debug information at the end of the response
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
	} else if removeHashtag(&text, "#message-history") {
		// Parse message-history without number (default to 10)
		options.MessageHistory = 10
	}

	// Parse model: #model-gpt-4o
	modelRe := regexp.MustCompile(`#model-([\w.-]+)`)
	if matches := modelRe.FindStringSubmatch(text); len(matches) > 1 {
		options.Model = matches[1]
		text = modelRe.ReplaceAllString(text, "")
	}

	// Parse reasoning effort
	reasoningMap := map[string]string{
		"#high-thinking":    "high",
		"#medium-thinking":  "medium",
		"#minimal-thinking": "minimal",
		"#no-thinking":      "none",
	}

	for hashtag, effort := range reasoningMap {
		if removeHashtag(&text, hashtag) {
			options.ReasoningEffort = effort
			break
		}
	}

	// Parse no-streaming option
	options.NoStreaming = removeHashtag(&text, "#no-streaming")

	// Parse no-thread option
	options.NoThread = removeHashtag(&text, "#no-thread")

	// Parse debug option
	options.Debug = removeHashtag(&text, "#debug")

	// Clean up extra whitespace
	cleanText = strings.Join(strings.Fields(text), " ")

	return cleanText, options
}
