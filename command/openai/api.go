package openai

import (
	"bytes"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

const (
	apiHost                  = "https://api.openai.com"
	apiCompletionURL         = "/v1/chat/completions"
	apiDalleGenerateImageURL = "/v1/images/generations"
)

const (
	roleUser      = "user"
	roleSystem    = "system"
	roleAssistant = "assistant"
)

// we don't use our default clients.HttpClient as we need longer timeouts...
var httpClient = http.Client{
	Timeout: 60 * time.Second,
}

func doRequest(cfg Config, apiEndpoint string, data []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", cfg.APIHost+apiEndpoint, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	return httpClient.Do(req)
}

// https://platform.openai.com/docs/api-reference/chat
type ChatRequest struct {
	Model            string        `json:"model"`
	Messages         []ChatMessage `json:"messages"`
	Temperature      float32       `json:"temperature,omitempty"`
	TopP             float32       `json:"top_p,omitempty"`
	N                int           `json:"n,omitempty"`
	Stop             []string      `json:"stop,omitempty"`
	Stream           bool          `json:"stream,omitempty"`
	MaxTokens        int           `json:"max_tokens,omitempty"`
	PresencePenalty  float32       `json:"presence_penalty,omitempty"`
	FrequencyPenalty float32       `json:"frequency_penalty,omitempty"`
	User             string        `json:"user,omitempty"`
	Seed             string        `json:"seed,omitempty"`
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

func (r ChatResponse) GetMessage() ChatMessage {
	return r.Choices[0].Message
}

func (r ChatResponse) GetDelta() ChatMessage {
	return r.Choices[0].Delta
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
	Delta        ChatMessage `json:"delta"`
}

/*
	{
	    "model": "dall-e-3",
	    "prompt": "a white siamese cat",
	    "n": 1,
	    "size": "1024x1024"
	  }
*/
type DalleRequest struct {
	Model   string `json:"model"`
	Quality string `json:"quality,omitempty"`
	Prompt  string `json:"prompt"`
	N       int    `json:"n"`
	Size    string `json:"size"`
}

/*
	{
	  "created": 1700233554,
	  "data": [
	    {
	      "url": "https://XXXX"
	    }
	  ]
	}

or:

	{
	  "error": {
	    "code": "invalid_size",
	    "message": "The size is not supported by this model.",
	    "param": null,
	    "type": "invalid_request_error"
	  }
	}
*/
type DalleResponse struct {
	Data  []DalleResponseImage `json:"data"`
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type DalleResponseImage struct {
	URL           string `json:"url"`
	RevisedPrompt string `json:"revised_prompt"`
}
