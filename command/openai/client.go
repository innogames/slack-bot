package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// todo: don't use our default client, we need longer timeouts...
var client *http.Client

func CallChatGPT(cfg Config, inputMessages []ChatMessage) (*ChatResponse, error) {
	jsonData, _ := json.Marshal(ChatRequest{
		Model:       cfg.Model,
		Temperature: cfg.Temperature,
		Messages:    inputMessages,
	})

	fmt.Println(string(jsonData))
	req, err := http.NewRequest("POST", cfg.APIHost+apiCompletionURL, bytes.NewBuffer(jsonData))
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
