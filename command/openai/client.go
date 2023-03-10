package openai

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

const (
	chatGPTAPIURL = "https://api.openai.com/v1/chat/completions"
	defaultModel  = "gpt-3.5-turbo"
)

const (
	roleUser      = "user"
	roleAssistant = "assistant"
)

// todo: don't use our default client, we need longer timeouts
var client http.Client

func CallChatGPT(cfg Config, inputMessages []ChatMessage) (*ChatResponse, error) {
	requestData := ChatRequest{
		Model:    defaultModel,
		Messages: inputMessages,
	}

	// Send the request to the API
	jsonData, err := json.Marshal(requestData)

	log.Info("openai", string(jsonData))

	req, err := http.NewRequest("POST", chatGPTAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Info("openai", string(body))

	var chatResponse ChatResponse
	err = json.Unmarshal(body, &chatResponse)
	if err != nil {
		return nil, err
	}

	if err = chatResponse.GetError(); err != nil {
		return &chatResponse, err
	}

	return &chatResponse, nil
}
