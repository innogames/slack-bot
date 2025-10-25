package openai

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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
	"github.com/stretchr/testify/require"
)

type testRequest struct {
	inputJSON    string
	responseJSON string
	responseCode int
}

func startTestServer(t *testing.T, url string, requests []testRequest) (Config, *httptest.Server) {
	t.Helper()

	idx := 0

	openaiCfg := defaultConfig

	mux := http.NewServeMux()
	mux.HandleFunc(url, func(res http.ResponseWriter, req *http.Request) {
		expected := requests[idx]
		idx++

		givenInputJSON, _ := io.ReadAll(req.Body)

		if len(expected.inputJSON) > 0 {
			assert.JSONEq(t, expected.inputJSON, string(givenInputJSON))
		}

		res.WriteHeader(expected.responseCode)

		response := strings.ReplaceAll(expected.responseJSON, "{test_server}", openaiCfg.APIHost)
		res.Write([]byte(response))
	})

	server := httptest.NewServer(mux)

	openaiCfg.APIHost = server.URL
	openaiCfg.APIKey = "0815pass"
	openaiCfg.Model = "gpt-4o"

	return openaiCfg, server
}

func TestOpenai(t *testing.T) {
	// init memory based storage
	storage.InitStorage("")

	slackClient := mocks.NewSlackClient(t)
	base := bot.BaseCommand{SlackClient: slackClient}

	t.Run("Openai is not active", func(t *testing.T) {
		cfg := &config.Config{}
		commands := GetCommands(base, cfg)
		assert.Equal(t, 0, commands.Count())
	})

	t.Run("Start a new thread and reply", func(t *testing.T) {
		openaiCfg, ts := startTestServer(
			t,
			apiCompletionURL,
			[]testRequest{
				{
					`{"model":"gpt-4o","messages":[{"role":"system","content":"You are a helpful Slack bot. By default, keep your answer short and truthful"},{"role":"user","content":"whats 1+1?"}],"stream":true}`,
					`data: {"id":"chatcmpl-6tuxebSPdmd2IJpb8GrZXHiYXON6r","object":"chat.completion.chunk","created":1678785018,"model":"gpt-4o-0301","choices":[{"delta":{"role":"assistant"},"index":0,"finish_reason":null}]}

data: {"id":"chatcmpl-6tuxebSPdmd2IJpb8GrZXHiYXON6r","object":"chat.completion.chunk","created":1678785018,"model":"gpt-4o-0301","choices":[{"delta":{"content":"The answer "},"index":0,"finish_reason":null}]}

data: {"id":"chatcmpl-6tuxebSPdmd2IJpb8GrZXHiYXON6r","object":"chat.completion.chunk","created":1678785018,"model":"gpt-4o-0301","choices":[{"delta":{"content":"is 2"},"index":0,"finish_reason":null}]}

data: {"id":"chatcmpl-6tuxebSPdmd2IJpb8GrZXHiYXON6r","object":"chat.completion.chunk","created":1678785018,"model":"gpt-4o-0301","choices":[{"delta":{},"index":0,"finish_reason":"stop"}]}

data: [DONE]`,
					http.StatusOK,
				},
				{
					`{"model":"gpt-4o","messages":[{"role":"system","content":"You are a helpful Slack bot. By default, keep your answer short and truthful"},{"role":"user","content":"whats 1+1?"},{"role":"assistant","content":"The answer is 2"},{"role":"user","content":"whats 2+1?"}],"stream":true}`,
					`{
						 "id": "chatcmpl-6p9XYPYSTTRi0xEviKjjilqrWU2Ve",
						 "object": "chat.completion",
						 "created": 1677649420,
						 "model": "gpt-4o",
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

		openaiCfg.LogTexts = true
		cfg := &config.Config{}
		cfg.Set("openai", openaiCfg)

		commands := GetCommands(base, cfg)
		assert.Equal(t, 1, commands.Count())

		help := commands.GetHelp()
		assert.Len(t, help, 2)

		message := msg.Message{}
		message.Text = "openai whats 1+1?"
		message.Channel = "testchan"
		message.Timestamp = "1234"
		ref := message.MessageRef

		mocks.AssertReaction(slackClient, ":bulb:", ref)
		mocks.AssertReaction(slackClient, ":speech_balloon:", ref)
		mocks.AssertRemoveReaction(slackClient, ":bulb:", ref)
		mocks.AssertRemoveReaction(slackClient, ":speech_balloon:", ref)
		mocks.AssertSlackMessage(slackClient, ref, ":bulb: thinking...", mock.Anything)
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

		mocks.AssertReaction(slackClient, ":bulb:", message)
		mocks.AssertReaction(slackClient, ":speech_balloon:", message)
		mocks.AssertRemoveReaction(slackClient, ":bulb:", message)
		mocks.AssertRemoveReaction(slackClient, ":speech_balloon:", message)
		mocks.AssertSlackMessage(slackClient, message, ":bulb: thinking...", mock.Anything)

		actual = commands.Run(message)
		queue.WaitTillHavingNoQueuedMessage()
		assert.True(t, actual)
	})

	t.Run("test http error", func(t *testing.T) {
		// mock openai API
		openaiCfg, ts := startTestServer(
			t,
			apiCompletionURL,
			[]testRequest{
				{
					`{"model":"gpt-4o","messages":[{"role":"user","content":"whats 1+1?"}],"stream":true}`,
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

		openaiCfg.InitialSystemMessage = ""
		cfg := &config.Config{}
		cfg.Set("openai", openaiCfg)
		commands := GetCommands(base, cfg)

		message := msg.Message{}
		message.Text = "openai whats 1+1?"
		ref := message.MessageRef

		mocks.AssertReaction(slackClient, ":bulb:", ref)
		mocks.AssertRemoveReaction(slackClient, ":bulb:", ref)
		mocks.AssertSlackMessage(slackClient, ref, ":bulb: thinking...", mock.Anything)
		mocks.AssertSlackMessage(slackClient, ref, "Incorrect API key provided: sk-1234**************************************567.", mock.Anything, mock.Anything)

		actual := commands.Run(message)
		queue.WaitTillHavingNoQueuedMessage()
		assert.True(t, actual)
	})

	t.Run("use as fallback", func(t *testing.T) {
		// mock openai API
		openaiCfg, ts := startTestServer(
			t,
			apiCompletionURL,

			[]testRequest{
				{
					`{"model":"gpt-4o","messages":[{"role":"user","content":"whats 1+1?"}],"stream":true}`,
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

		openaiCfg.UseAsFallback = true
		openaiCfg.InitialSystemMessage = ""
		cfg := &config.Config{}
		cfg.Set("openai", openaiCfg)
		commands := GetCommands(base, cfg)

		message := msg.Message{}
		message.Text = "whats 1+1?"

		mocks.AssertReaction(slackClient, ":bulb:", message)
		mocks.AssertRemoveReaction(slackClient, ":bulb:", message)
		mocks.AssertSlackMessage(slackClient, message, ":bulb: thinking...", mock.Anything)
		mocks.AssertSlackMessage(slackClient, message, "Incorrect API key provided: sk-1234**************************************567.", mock.Anything, mock.Anything)

		actual := commands.Run(message)
		queue.WaitTillHavingNoQueuedMessage()
		assert.True(t, actual)
	})

	t.Run("render template", func(t *testing.T) {
		// mock openai API
		openaiCfg, ts := startTestServer(
			t,
			apiCompletionURL,
			[]testRequest{
				{
					`{"model":"gpt-4o","messages":[{"role":"user","content":"whats 1+1?"}]}`,
					`{
						 "id": "chatcmpl-6p9XYPYSTTRi0xEviKjjilqrWU2Ve",
						 "object": "chat.completion",
						 "created": 1677649420,
						 "model": "gpt-4o",
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

		command := openaiCommand{base, openaiCfg}

		util.RegisterFunctions(command.GetTemplateFunction())
		tpl, err := util.CompileTemplate(`{{ openai "whats 1+1?"}}`)
		require.NoError(t, err)

		res, err := util.EvalTemplate(tpl, util.Parameters{})
		require.NoError(t, err)

		assert.Equal(t, "The answer is 2", res)
	})

	t.Run("Write within a new thread", func(t *testing.T) {
		openaiCfg, ts := startTestServer(
			t,
			apiCompletionURL,
			[]testRequest{
				{
					`{"model":"gpt-4o","messages":[{"role":"system","content":"You are a helpful Slack bot. By default, keep your answer short and truthful"},{"role":"system","content":"This is a Slack bot receiving a slack thread s context, using slack user ids as identifiers. Please use user mentions in the format \u003c@U123456\u003e"},{"role":"user","content":"User \u003c@U1234\u003e wrote: thread message 1"},{"role":"user","content":"whats 1+1?"}],"stream":true}`,
					`data: {"id":"chatcmpl-6tuxebSPdmd2IJpb8GrZXHiYXON6r","object":"chat.completion.chunk","created":1678785018,"model":"gpt-4o-0301","choices":[{"delta":{"role":"assistant"},"index":0,"finish_reason":null}]}

data: {"id":"chatcmpl-6tuxebSPdmd2IJpb8GrZXHiYXON6r","object":"chat.completion.chunk","created":1678785018,"model":"gpt-4o-0301","choices":[{"delta":{"content":"Jolo!"},"index":0,"finish_reason":null}]}

data: {"id":"chatcmpl-6tuxebSPdmd2IJpb8GrZXHiYXON6r","object":"chat.completion.chunk","created":1678785018,"model":"gpt-4o-0301","choices":[{"delta":{},"index":0,"finish_reason":"stop"}]}

data: [DONE]`,
					http.StatusOK,
				},
			},
		)
		defer ts.Close()

		openaiCfg.LogTexts = true
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

		mocks.AssertReaction(slackClient, ":bulb:", ref)
		mocks.AssertReaction(slackClient, ":speech_balloon:", ref)
		mocks.AssertRemoveReaction(slackClient, ":bulb:", ref)
		mocks.AssertRemoveReaction(slackClient, ":speech_balloon:", ref)
		mocks.AssertSlackMessage(slackClient, ref, ":bulb: thinking...", mock.Anything)
		mocks.AssertSlackMessage(slackClient, ref, "Jolo!", mock.Anything, mock.Anything)

		actual = commands.Run(message)
		queue.WaitTillHavingNoQueuedMessage()
		assert.True(t, actual)
	})

	t.Run("Other thread given", func(t *testing.T) {
		openaiCfg, ts := startTestServer(
			t,
			apiCompletionURL,
			[]testRequest{
				{
					`{"model":"dummy-test","messages":[{"role":"user","content":"User \u003c@U1234\u003e wrote: i had a great weekend"},{"role":"user","content":"summarize this thread "}],"stream":true}`,
					`data: {"id":"chatcmpl-6tuxebSPdmd2IJpb8GrZXHiYXON6r","object":"chat.completion.chunk","created":1678785018,"model":"gpt-4o-0301","choices":[{"delta":{"role":"assistant"},"index":0,"finish_reason":null}]}

data: {"id":"chatcmpl-6tuxebSPdmd2IJpb8GrZXHiYXON6r","object":"chat.completion.chunk","created":1678785018,"model":"gpt-4o-0301","choices":[{"delta":{"content":"Jolo!"},"index":0,"finish_reason":null}]}

data: {"id":"chatcmpl-6tuxebSPdmd2IJpb8GrZXHiYXON6r","object":"chat.completion.chunk","created":1678785018,"model":"gpt-4o-0301","choices":[{"delta":{},"index":0,"finish_reason":"stop"}]}

data: [DONE]`,
					http.StatusOK,
				},
			},
		)
		defer ts.Close()

		openaiCfg.InitialSystemMessage = ""
		openaiCfg.Model = "dummy-test"
		cfg := &config.Config{}
		cfg.Set("openai", openaiCfg)

		commands := GetCommands(base, cfg)

		message := msg.Message{}
		message.Text = "openai summarize this thread <https://foobar.slack.com/archives/chan1234/p1694166741200139>"
		message.Channel = "testchan"
		message.Timestamp = "1234"

		threadMessage1 := slack.Message{}
		threadMessage1.User = "U1234"
		threadMessage1.Channel = "chan1234"
		threadMessage1.Text = "i had a great weekend"
		threadMessage2 := slack.Message{}
		threadMessage2.User = "U1234"
		threadMessage2.Channel = "chan1234"
		threadMessage2.Text = "Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore asdasd adasd asdasd, Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore asdasd adasd asdasd Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut labore et dolore asdasd adasd asdasd"

		ref := msg.MessageRef{
			Channel: "chan1234",
			Thread:  "1694166741.200139",
		}

		threadMessages := []slack.Message{threadMessage1, threadMessage2}
		slackClient.On("GetThreadMessages", ref).Once().Return(threadMessages, nil)

		mocks.AssertReaction(slackClient, ":bulb:", message.MessageRef)
		mocks.AssertReaction(slackClient, ":speech_balloon:", message.MessageRef)
		mocks.AssertRemoveReaction(slackClient, ":bulb:", message.MessageRef)
		mocks.AssertRemoveReaction(slackClient, ":speech_balloon:", message.MessageRef)
		mocks.AssertSlackMessage(slackClient, message.MessageRef, "Note: The token length of 100 exceeded! 1 messages were not sent", mock.Anything)
		mocks.AssertSlackMessage(slackClient, message.MessageRef, ":bulb: thinking...", mock.Anything)
		mocks.AssertSlackMessage(slackClient, message.MessageRef, "Jolo!", mock.Anything, mock.Anything)

		actual := commands.Run(message)
		time.Sleep(time.Millisecond * 10)
		queue.WaitTillHavingNoQueuedMessage()
		assert.True(t, actual)
	})

	t.Run("Test no-streaming hashtag", func(t *testing.T) {
		openaiCfg, ts := startTestServer(
			t,
			apiCompletionURL,
			[]testRequest{
				{
					`{"model":"gpt-4o","messages":[{"role":"system","content":"You are a helpful Slack bot. By default, keep your answer short and truthful"},{"role":"user","content":"whats 1+1?"}]}`,
					`{
						 "id": "chatcmpl-6p9XYPYSTTRi0xEviKjjilqrWU2Ve",
						 "object": "chat.completion",
						 "created": 1677649420,
						 "model": "gpt-4o",
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
			},
		)
		defer ts.Close()

		openaiCfg.LogTexts = true
		cfg := &config.Config{}
		cfg.Set("openai", openaiCfg)

		commands := GetCommands(base, cfg)

		message := msg.Message{}
		message.Text = "openai #no-streaming whats 1+1?"
		message.Channel = "testchan"
		message.Timestamp = "1234"
		ref := message.MessageRef

		mocks.AssertReaction(slackClient, ":bulb:", ref)
		mocks.AssertReaction(slackClient, ":speech_balloon:", ref)
		mocks.AssertRemoveReaction(slackClient, ":bulb:", ref)
		mocks.AssertRemoveReaction(slackClient, ":speech_balloon:", ref)
		mocks.AssertSlackMessage(slackClient, ref, ":bulb: thinking...", mock.Anything)
		mocks.AssertSlackMessage(slackClient, ref, "The answer is 2", mock.Anything, mock.Anything)

		actual := commands.Run(message)
		queue.WaitTillHavingNoQueuedMessage()
		assert.True(t, actual)
	})

	t.Run("Test no-thread hashtag on channel level", func(t *testing.T) {
		openaiCfg, ts := startTestServer(
			t,
			apiCompletionURL,
			[]testRequest{
				{
					`{"model":"gpt-4o","messages":[{"role":"system","content":"You are a helpful Slack bot. By default, keep your answer short and truthful"},{"role":"user","content":"quick question"}],"stream":true}`,
					`data: {"id":"chatcmpl-6tuxebSPdmd2IJpb8GrZXHiYXON6r","object":"chat.completion.chunk","created":1678785018,"model":"gpt-4o-0301","choices":[{"delta":{"role":"assistant"},"index":0,"finish_reason":null}]}

data: {"id":"chatcmpl-6tuxebSPdmd2IJpb8GrZXHiYXON6r","object":"chat.completion.chunk","created":1678785018,"model":"gpt-4o-0301","choices":[{"delta":{"content":"Sure, what is it?"},"index":0,"finish_reason":null}]}

data: {"id":"chatcmpl-6tuxebSPdmd2IJpb8GrZXHiYXON6r","object":"chat.completion.chunk","created":1678785018,"model":"gpt-4o-0301","choices":[{"delta":{},"index":0,"finish_reason":"stop"}]}

data: [DONE]`,
					http.StatusOK,
				},
			},
		)
		defer ts.Close()

		cfg := &config.Config{}
		cfg.Set("openai", openaiCfg)

		commands := GetCommands(base, cfg)

		message := msg.Message{}
		message.Text = "openai #no-thread quick question"
		message.Channel = "testchan"
		message.Timestamp = "1234"
		ref := message.MessageRef

		mocks.AssertReaction(slackClient, ":bulb:", ref)
		mocks.AssertReaction(slackClient, ":speech_balloon:", ref)
		mocks.AssertRemoveReaction(slackClient, ":bulb:", ref)
		mocks.AssertRemoveReaction(slackClient, ":speech_balloon:", ref)
		// When #no-thread is used at channel level, message should be sent without thread
		slackClient.On("SendMessage", ref, ":bulb: thinking...").Return("1234.1").Once()
		slackClient.On("SendMessage", ref, "Sure, what is it?", mock.Anything).Return("").Once()

		actual := commands.Run(message)
		queue.WaitTillHavingNoQueuedMessage()
		assert.True(t, actual)
	})

	t.Run("Test no-thread hashtag within a thread (should be ignored)", func(t *testing.T) {
		openaiCfg, ts := startTestServer(
			t,
			apiCompletionURL,
			[]testRequest{
				{
					`{"model":"gpt-4o","messages":[{"role":"system","content":"You are a helpful Slack bot. By default, keep your answer short and truthful"},{"role":"system","content":"This is a Slack bot receiving a slack thread s context, using slack user ids as identifiers. Please use user mentions in the format \u003c@U123456\u003e"},{"role":"user","content":"User \u003c@U1234\u003e wrote: previous message"},{"role":"user","content":"another question"}],"stream":true}`,
					`data: {"id":"chatcmpl-6tuxebSPdmd2IJpb8GrZXHiYXON6r","object":"chat.completion.chunk","created":1678785018,"model":"gpt-4o-0301","choices":[{"delta":{"role":"assistant"},"index":0,"finish_reason":null}]}

data: {"id":"chatcmpl-6tuxebSPdmd2IJpb8GrZXHiYXON6r","object":"chat.completion.chunk","created":1678785018,"model":"gpt-4o-0301","choices":[{"delta":{"content":"Yes, I can help"},"index":0,"finish_reason":null}]}

data: {"id":"chatcmpl-6tuxebSPdmd2IJpb8GrZXHiYXON6r","object":"chat.completion.chunk","created":1678785018,"model":"gpt-4o-0301","choices":[{"delta":{},"index":0,"finish_reason":"stop"}]}

data: [DONE]`,
					http.StatusOK,
				},
			},
		)
		defer ts.Close()

		cfg := &config.Config{}
		cfg.Set("openai", openaiCfg)

		commands := GetCommands(base, cfg)

		message := msg.Message{}
		message.Text = "openai #no-thread another question"
		message.Channel = "testchan"
		message.Timestamp = "5678"
		message.Thread = "1234" // Already in a thread
		ref := message.MessageRef

		// Mock the thread history load
		threadMessage := slack.Message{}
		threadMessage.User = "U1234"
		threadMessage.Text = "previous message"
		threadMessages := []slack.Message{threadMessage}

		threadRef := msg.MessageRef{
			Channel:   "testchan",
			Thread:    "1234",
			Timestamp: "5678",
		}
		slackClient.On("GetThreadMessages", threadRef).Once().Return(threadMessages, nil)

		mocks.AssertReaction(slackClient, ":bulb:", ref)
		mocks.AssertReaction(slackClient, ":speech_balloon:", ref)
		mocks.AssertRemoveReaction(slackClient, ":bulb:", ref)
		mocks.AssertRemoveReaction(slackClient, ":speech_balloon:", ref)
		// When already in a thread, #no-thread should be ignored and message should still be in thread
		mocks.AssertSlackMessage(slackClient, ref, ":bulb: thinking...", mock.Anything)
		mocks.AssertSlackMessage(slackClient, ref, "Yes, I can help", mock.Anything, mock.Anything)

		actual := commands.Run(message)
		queue.WaitTillHavingNoQueuedMessage()
		assert.True(t, actual)
	})
}
