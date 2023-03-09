package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const chatGPTAPIURL = "https://api.openai.com/v1/chat/completions"
const defaultModel = "gpt-3.5-turbo"

const (
	roleUser      = "user"
	roleAssistant = "assistant"
)

// todo: don't use our default client, we need longer timeouts
var client http.Client

func CallChatGPT(cfg OpenAIConfig, inputMessages []ChatMessage) (*ChatResponse, error) {
	requestData := ChatRequest{
		Model:    defaultModel,
		Messages: inputMessages,
	}

	// Send the request to the API
	jsonData, err := json.Marshal(requestData)

	fmt.Println(string(jsonData))

	req, err := http.NewRequest("POST", chatGPTAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.ApiKey)

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

	fmt.Println(string(body))

	var chatResponse ChatResponse
	err = json.Unmarshal(body, &chatResponse)
	if err != nil {
		return nil, err
	}

	if err = chatResponse.GetError(); err != nil {
		return &chatResponse, err
	}

	fmt.Println("resp")
	fmt.Println(string(body))

	return &chatResponse, nil
}
