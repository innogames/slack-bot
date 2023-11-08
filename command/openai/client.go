package openai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

// we don't use our default clients.HttpClient as we need longer timeouts...
var client http.Client

func CallChatGPT(cfg Config, inputMessages []ChatMessage, stream bool) (<-chan string, error) {
	jsonData, _ := json.Marshal(ChatRequest{
		Model:       cfg.Model,
		Temperature: cfg.Temperature,
		Seed:        cfg.Seed,
		MaxTokens:   cfg.MaxTokens,
		Stream:      stream,
		Messages:    inputMessages,
	})

	req, err := http.NewRequest("POST", cfg.APIHost+apiCompletionURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	messageUpdates := make(chan string, 2)

	// return a chan of all message updates here and listen here in the background in the event stream
	go func() {
		defer close(messageUpdates)

		resp, err := client.Do(req)
		if err != nil {
			messageUpdates <- err.Error()
			return
		}
		defer resp.Body.Close()

		// some error occurred: we don't have an event stream but a single ChatResponse with an error
		if !stream || resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)

			var chatResponse ChatResponse
			err = json.Unmarshal(body, &chatResponse)
			if err != nil {
				messageUpdates <- err.Error()
				return
			}

			if err = chatResponse.GetError(); err != nil {
				messageUpdates <- chatResponse.GetError().Error()
				return
			}

			if message := chatResponse.GetMessage().Content; message != "" {
				messageUpdates <- message
				return
			}

			return
		}

		// each line contains a delta of the message, so one new token
		fileScanner := bufio.NewScanner(resp.Body)
		fileScanner.Split(bufio.ScanLines)
		for fileScanner.Scan() {
			line := fileScanner.Text()
			if _, deltaJSON, found := strings.Cut(line, "data: "); found {
				if deltaJSON == "[DONE]" {
					// end of event stream
					return
				}

				var delta ChatResponse
				err = json.Unmarshal([]byte(deltaJSON), &delta)
				if err != nil {
					log.Warnf("openai error in json: %s (json: %s)", err, deltaJSON)
					continue
				}

				if deltaContent := delta.GetDelta().Content; deltaContent != "" {
					messageUpdates <- deltaContent
				}
			}
		}
	}()

	return messageUpdates, nil
}
