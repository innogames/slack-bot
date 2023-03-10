package openai

import "github.com/pkg/errors"

// https://github.com/otiai10/openaigo/blob/main/chat.go
// https://platform.openai.com/docs/guides/chat/chat-completions-beta
// https://platform.openai.com/docs/api-reference/chat
type ChatRequest struct {
	Model            string         `json:"model"`
	Messages         []ChatMessage  `json:"messages"`
	Temperature      float32        `json:"temperature,omitempty"`
	TopP             float32        `json:"top_p,omitempty"`
	N                int            `json:"n,omitempty"`
	Stop             []string       `json:"stop,omitempty"`
	MaxTokens        int            `json:"max_tokens,omitempty"`
	PresencePenalty  float32        `json:"presence_penalty,omitempty"`
	FrequencyPenalty float32        `json:"frequency_penalty,omitempty"`
	LogitBias        map[string]int `json:"logit_bias,omitempty"`
	User             string         `json:"user,omitempty"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Choices []ChatChoice `json:"choices"`
	Error   struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func (r ChatResponse) GetRecentMessage() string {
	return r.Choices[len(r.Choices)-1].Message.Content
}

func (r ChatResponse) GetError() error {
	if r.Error.Message == "" {
		return nil
	}

	return errors.New(r.Error.Message)
}

type ChatChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}