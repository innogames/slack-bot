package openai

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/storage"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/innogames/slack-bot/v2/command/queue"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/slack-go/slack"
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
	// init memory based storage
	storage.InitStorage("")

	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	t.Run("Openai is not active", func(t *testing.T) {
		cfg := &config.Config{}
		commands := GetCommands(base, cfg)
		assert.Equal(t, 0, commands.Count())
	})

	t.Run("Start a new thread and reply", func(t *testing.T) {
		ts := startTestServer(
			t,
			[]testRequest{
				{
					`{"model":"gpt-3.5-turbo","messages":[{"role":"system","content":"You are a helpful Slack bot. By default, keep your answer short and truthful"},{"role":"user","content":"whats 1+1?"}],"stream":true}`,
					`data: {"id":"chatcmpl-6tuxebSPdmd2IJpb8GrZXHiYXON6r","object":"chat.completion.chunk","created":1678785018,"model":"gpt-3.5-turbo-0301","choices":[{"delta":{"role":"assistant"},"index":0,"finish_reason":null}]}

data: {"id":"chatcmpl-6tuxebSPdmd2IJpb8GrZXHiYXON6r","object":"chat.completion.chunk","created":1678785018,"model":"gpt-3.5-turbo-0301","choices":[{"delta":{"content":"The answer "},"index":0,"finish_reason":null}]}

data: {"id":"chatcmpl-6tuxebSPdmd2IJpb8GrZXHiYXON6r","object":"chat.completion.chunk","created":1678785018,"model":"gpt-3.5-turbo-0301","choices":[{"delta":{"content":"is 2"},"index":0,"finish_reason":null}]}

data: {"id":"chatcmpl-6tuxebSPdmd2IJpb8GrZXHiYXON6r","object":"chat.completion.chunk","created":1678785018,"model":"gpt-3.5-turbo-0301","choices":[{"delta":{},"index":0,"finish_reason":"stop"}]}

data: [DONE]`,
					http.StatusOK,
				},
				{
					`{"model":"gpt-3.5-turbo","messages":[{"role":"system","content":"You are a helpful Slack bot. By default, keep your answer short and truthful"},{"role":"user","content":"whats 1+1?"},{"role":"assistant","content":"The answer is 2"},{"role":"user","content":"whats 2+1?"}],"stream":true}`,
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
		ref := message.MessageRef

		mocks.AssertReaction(slackClient, ":coffee:", ref)
		mocks.AssertRemoveReaction(slackClient, ":coffee:", ref)
		mocks.AssertSlackMessage(slackClient, ref, "...", mock.Anything)
		mocks.AssertSlackMessage(slackClient, ref, "The answer ", mock.Anything, mock.Anything)
		mocks.AssertSlackMessage(slackClient, ref, "The answer is 2", mock.Anything, mock.Anything)

		actual := commands.Run(message)
		queue.WaitTillHavingNoQueuedMessage()
		assert.True(t, actual)

		// test reply in different context -> nothing
		message = msg.Message{}
		message.Text = "whats 1+1?"
		message.Channel = "testchan"
		message.Thread = "4321"

		actual = commands.Run(message)
		queue.WaitTillHavingNoQueuedMessage()
		assert.False(t, actual)

		// test reply in same context -> ask openai with history
		message = msg.Message{}
		message.Text = "whats 2+1?"
		message.Channel = "testchan"
		message.Thread = "1234"

		mocks.AssertReaction(slackClient, ":coffee:", message)
		mocks.AssertRemoveReaction(slackClient, ":coffee:", message)
		mocks.AssertSlackMessage(slackClient, message, "...", mock.Anything)
		mocks.AssertSlackMessage(slackClient, message, "The answer is 3", mock.Anything)

		actual = commands.Run(message)
		queue.WaitTillHavingNoQueuedMessage()
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
		ref := message.MessageRef

		mocks.AssertReaction(slackClient, ":coffee:", ref)
		mocks.AssertRemoveReaction(slackClient, ":coffee:", ref)
		mocks.AssertSlackMessage(slackClient, ref, "...", mock.Anything)
		mocks.AssertSlackMessage(slackClient, ref, "Incorrect API key provided: sk-1234**************************************567.", mock.Anything, mock.Anything)

		actual := commands.Run(message)
		queue.WaitTillHavingNoQueuedMessage()
		assert.True(t, actual)
	})

	t.Run("use as fallback", func(t *testing.T) {
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
			APIHost:       ts.URL,
			APIKey:        "0815pass",
			UseAsFallback: true,
		}
		cfg := &config.Config{}
		cfg.Set("openai", openaiCfg)
		commands := GetCommands(base, cfg)

		message := msg.Message{}
		message.Text = "whats 1+1?"

		mocks.AssertReaction(slackClient, ":coffee:", message)
		mocks.AssertRemoveReaction(slackClient, ":coffee:", message)
		mocks.AssertSlackMessage(slackClient, message, "...", mock.Anything)
		mocks.AssertSlackMessage(slackClient, message, "Incorrect API key provided: sk-1234**************************************567.", mock.Anything, mock.Anything)

		actual := commands.Run(message)
		queue.WaitTillHavingNoQueuedMessage()
		assert.True(t, actual)
	})

	t.Run("render template", func(t *testing.T) {
		// mock openai API
		ts := startTestServer(
			t,
			[]testRequest{
				{
					`{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"whats 1+1?"}]}`,
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
							  "content": "\n\nThe answer is 2"},
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
		command := chatGPTCommand{base, openaiCfg}

		util.RegisterFunctions(command.GetTemplateFunction())
		tpl, err := util.CompileTemplate(`{{ openai "whats 1+1?"}}`)
		assert.Nil(t, err)

		res, err := util.EvalTemplate(tpl, util.Parameters{})
		assert.Nil(t, err)

		assert.Equal(t, "The answer is 2", res)
	})

	t.Run("Write within a new thread", func(t *testing.T) {
		ts := startTestServer(
			t,
			[]testRequest{
				{
					`{"model":"gpt-3.5-turbo","messages":[{"role":"system","content":"You are a helpful Slack bot. By default, keep your answer short and truthful"},{"role":"system","content":"This is a Slack bot receiving a slack thread s context, using slack user ids as identifiers. Please use user mentions in the format \u003c@U123456\u003e"},{"role":"user","content":"User \u003c@U1234\u003e wrote: thread message 1"},{"role":"user","content":"whats 1+1?"}],"stream":true}`,
					`data: {"id":"chatcmpl-6tuxebSPdmd2IJpb8GrZXHiYXON6r","object":"chat.completion.chunk","created":1678785018,"model":"gpt-3.5-turbo-0301","choices":[{"delta":{"role":"assistant"},"index":0,"finish_reason":null}]}

data: {"id":"chatcmpl-6tuxebSPdmd2IJpb8GrZXHiYXON6r","object":"chat.completion.chunk","created":1678785018,"model":"gpt-3.5-turbo-0301","choices":[{"delta":{"content":"Jolo!"},"index":0,"finish_reason":null}]}

data: {"id":"chatcmpl-6tuxebSPdmd2IJpb8GrZXHiYXON6r","object":"chat.completion.chunk","created":1678785018,"model":"gpt-3.5-turbo-0301","choices":[{"delta":{},"index":0,"finish_reason":"stop"}]}

data: [DONE]`,
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

		message := msg.Message{}
		message.Text = "openai whats 1+1?"
		message.Channel = "testchan"
		message.Thread = "12345"
		message.Timestamp = "1234"
		ref := message.MessageRef

		// first with an error
		mocks.AssertError(slackClient, ref, "can't load thread messages: openai not reachable")
		slackClient.On("GetThreadMessages", ref).Once().Return([]slack.Message{}, errors.New("openai not reachable"))
		actual := commands.Run(message)
		queue.WaitTillHavingNoQueuedMessage()
		assert.True(t, actual)

		// then a successful attempt
		threadMessage := slack.Message{}
		threadMessage.User = "U1234"
		threadMessage.Text = "thread message 1"

		threadMessages := []slack.Message{threadMessage}
		slackClient.On("GetThreadMessages", ref).Once().Return(threadMessages, nil)

		mocks.AssertReaction(slackClient, ":coffee:", ref)
		mocks.AssertRemoveReaction(slackClient, ":coffee:", ref)
		mocks.AssertSlackMessage(slackClient, ref, "...", mock.Anything)
		mocks.AssertSlackMessage(slackClient, ref, "Jolo!", mock.Anything, mock.Anything)

		actual = commands.Run(message)
		queue.WaitTillHavingNoQueuedMessage()
		assert.True(t, actual)
	})
}
