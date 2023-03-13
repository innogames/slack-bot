package openai

// TODO cleanup before merge

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/innogames/slack-bot/v2/bot/storage"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type testRequest struct {
	inputJSON    string
	responseJSON string
	responseCode int
}

func startTestServer(t *testing.T, requests []testRequest) *httptest.Server {
	t.Helper()

	idx := 0

	mux := http.NewServeMux()
	mux.HandleFunc(apiCompletionURL, func(res http.ResponseWriter, req *http.Request) {
		expected := requests[idx]
		idx++

		givenInputJSON, _ := io.ReadAll(req.Body)

		assert.Equal(t, expected.inputJSON, string(givenInputJSON))

		res.WriteHeader(expected.responseCode)
		res.Write([]byte(expected.responseJSON))
	})

	return httptest.NewServer(mux)
}

func TestOpenai(t *testing.T) {
	storage.InitStorage("")

	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	t.Run("Openai is not active", func(t *testing.T) {
		cfg := &config.Config{}
		commands := GetCommands(base, cfg)
		assert.Equal(t, 0, commands.Count())
	})

	t.Run("Default flow", func(t *testing.T) {
		ts := startTestServer(
			t,
			[]testRequest{
				{
					`{"model":"gpt-3.5-turbo","messages":[{"role":"system","content":"You are a helpful Slack bot. By default, keep your answer short and truthful"},{"role":"user","content":"whats 1+1?"}],"stream":true}`,
					`{
						 "id": "chatcmpl-6p9XYPYSTTRi0xEviKjjilqrWU2Ve",
						 "object": "chat.completion",
						 "created": 1677649420,
						 "model": "gpt-3.5-turbo",
						 "usage": {"prompt_tokens": 56, "completion_tokens": 31, "total_tokens": 87},
						 "choices": [
						   {
							"message": {
							  "role": "assistant",
							  "content": "The answer is 2"},
							"finish_reason": "stop",
							"index": 0
						   }
						  ]
						}`,
					http.StatusOK,
				},
				{
					`{"model":"gpt-3.5-turbo","messages":[{"role":"system","content":"You are a helpful Slack bot. By default, keep your answer short and truthful"},{"role":"user","content":"whats 1+1?"},{"role":"user","content":""},{"role":"user","content":"whats 2+1?"}],"stream":true}`,
					`{
						 "id": "chatcmpl-6p9XYPYSTTRi0xEviKjjilqrWU2Ve",
						 "object": "chat.completion",
						 "created": 1677649420,
						 "model": "gpt-3.5-turbo",
						 "usage": {"prompt_tokens": 56, "completion_tokens": 31, "total_tokens": 87},
						 "choices": [
						   {
							"message": {
							  "role": "assistant",
							  "content": "The answer is 3"},
							"finish_reason": "stop",
							"index": 0
						   }
						  ]
						}`,
					http.StatusOK,
				},
			},
		)
		defer ts.Close()

		openaiCfg := defaultConfig
		openaiCfg.APIHost = ts.URL
		openaiCfg.APIKey = "0815pass"
		cfg := &config.Config{}
		cfg.Set("openai", openaiCfg)

		commands := GetCommands(base, cfg)
		assert.Equal(t, 1, commands.Count())

		help := commands.GetHelp()
		assert.Equal(t, 1, len(help))

		message := msg.Message{}
		message.Text = "openai whats 1+1?"
		message.Channel = "testchan"
		message.Timestamp = "1234"

		mocks.AssertReaction(slackClient, ":coffee:", message)
		mocks.AssertRemoveReaction(slackClient, ":coffee:", message)
		slackClient.On("SendMessage", message, "...", mock.Anything).Once().Return("")
		slackClient.On("SendMessage", message, "The answer is 2", mock.Anything).Once().Return("")

		actual := commands.Run(message)
		time.Sleep(time.Millisecond * 500) // todo wait for go routines
		assert.True(t, actual)

		// test reply in different context -> nothing
		message = msg.Message{}
		message.Text = "whats 1+1?"
		message.Channel = "testchan"
		message.Thread = "4321"

		actual = commands.Run(message)
		time.Sleep(time.Millisecond * 500) // todo wait for go routines
		assert.False(t, actual)

		// test reply in same context -> ask openai with history
		message = msg.Message{}
		message.Text = "whats 2+1?"
		message.Channel = "testchan"
		message.Thread = "1234"

		mocks.AssertReaction(slackClient, ":coffee:", message)
		mocks.AssertRemoveReaction(slackClient, ":coffee:", message)
		slackClient.On("SendMessage", message, "...", mock.Anything).Once().Return("")
		slackClient.On("SendMessage", message, "The answer is 3", mock.Anything).Once().Return("")

		actual = commands.Run(message)
		time.Sleep(time.Second * 2) // todo wait for go routines
		assert.True(t, actual)
	})

	t.Run("test http error", func(t *testing.T) {
		// mock openai API
		ts := startTestServer(
			t,
			[]testRequest{
				{
					`{"model":"","messages":[{"role":"user","content":"whats 1+1?"}],"stream":true}`,
					`{
					  "error": {
						"code": "invalid_api_key",
						"message": "Incorrect API key provided: sk-1234**************************************567.",
						"type": "invalid_request_error"
					  }
					}`,
					http.StatusUnauthorized,
				},
			},
		)
		defer ts.Close()

		openaiCfg := Config{
			APIHost: ts.URL,
			APIKey:  "0815pass",
		}
		cfg := &config.Config{}
		cfg.Set("openai", openaiCfg)
		commands := GetCommands(base, cfg)

		message := msg.Message{}
		message.Text = "openai whats 1+1?"

		mocks.AssertReaction(slackClient, ":coffee:", message)
		mocks.AssertRemoveReaction(slackClient, ":coffee:", message)
		slackClient.On("SendMessage", message, "...", mock.Anything).Once().Return("")
		slackClient.On("SendMessage", message, "Incorrect API key provided: sk-1234**************************************567.", mock.Anything, mock.Anything).Once().Return("")

		actual := commands.Run(message)
		time.Sleep(100 * time.Millisecond)
		assert.True(t, actual)
	})
}
