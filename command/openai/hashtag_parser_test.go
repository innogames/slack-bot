package openai

import (
	"testing"
)

func TestParseHashtags(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedText    string
		expectedModel   string
		expectedReason  string
		expectedHistory int
	}{
		{
			name:          "No hashtags",
			input:         "What is Go?",
			expectedText:  "What is Go?",
			expectedModel: "",
		},
		{
			name:          "Model only",
			input:         "#model-gpt-4o What is Go?",
			expectedText:  "What is Go?",
			expectedModel: "gpt-4o",
		},
		{
			name:           "High thinking",
			input:          "#high-thinking Explain quantum computing",
			expectedText:   "Explain quantum computing",
			expectedReason: "high",
		},
		{
			name:            "Message history default",
			input:           "#message-history What was discussed?",
			expectedText:    "What was discussed?",
			expectedHistory: 10,
		},
		{
			name:            "Message history with count",
			input:           "#message-history-25 Summarize the conversation",
			expectedText:    "Summarize the conversation",
			expectedHistory: 25,
		},
		{
			name:            "Multiple hashtags",
			input:           "#model-o1 #high-thinking #message-history-15 Complex question",
			expectedText:    "Complex question",
			expectedModel:   "o1",
			expectedReason:  "high",
			expectedHistory: 15,
		},
		{
			name:           "No thinking",
			input:          "#no-thinking Quick answer please",
			expectedText:   "Quick answer please",
			expectedReason: "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanText, options := ParseHashtags(tt.input)

			if cleanText != tt.expectedText {
				t.Errorf("Expected text '%s', got '%s'", tt.expectedText, cleanText)
			}
			if options.Model != tt.expectedModel {
				t.Errorf("Expected model '%s', got '%s'", tt.expectedModel, options.Model)
			}
			if options.ReasoningEffort != tt.expectedReason {
				t.Errorf("Expected reasoning '%s', got '%s'", tt.expectedReason, options.ReasoningEffort)
			}
			if options.MessageHistory != tt.expectedHistory {
				t.Errorf("Expected history %d, got %d", tt.expectedHistory, options.MessageHistory)
			}
		})
	}
}
