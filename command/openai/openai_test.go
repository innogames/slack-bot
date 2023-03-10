package openai

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOpenai(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	t.Run("Openai is not active", func(t *testing.T) {
		cfg := &config.Config{}
		commands := GetCommands(base, cfg)
		assert.Equal(t, 0, commands.Count())
	})

	t.Run("Default flow", func(t *testing.T) {
		// mock openai API
		mux := http.NewServeMux()
		mux.HandleFunc(apiCompletionURL, func(res http.ResponseWriter, req *http.Request) {
			res.Write([]byte(`{
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
			}`))
		})
		ts := httptest.NewServer(mux)
		defer ts.Close()

		openaiCfg := Config{
			APIHost: ts.URL,
			APIKey:  "0815pass",
		}
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

		mocks.AssertReaction(slackClient, ":spiral_note_pad:", message)
		mocks.AssertReaction(slackClient, ":coffee:", message)
		mocks.AssertRemoveReaction(slackClient, ":coffee:", message)
		slackClient.On("SendMessage", message, "The answer is 2", mock.Anything).Once().Return("")

		actual := commands.Run(message)
		assert.True(t, actual)

		// test reply in different context -> nothing
		message = msg.Message{}
		message.Text = "whats 1+1?"
		message.Channel = "testchan"
		message.Thread = "4321"

		actual = commands.Run(message)
		assert.False(t, actual)

		// test reply in same context -> ask openai with history
		message = msg.Message{}
		message.Text = "whats 2+1?"
		message.Channel = "testchan"
		message.Thread = "1234"

		mocks.AssertReaction(slackClient, ":coffee:", message)
		mocks.AssertRemoveReaction(slackClient, ":coffee:", message)
		slackClient.On("SendMessage", message, "The answer is 2", mock.Anything).Once().Return("")

		actual = commands.Run(message)
		assert.True(t, actual)
	})

	t.Run("test http error", func(t *testing.T) {
		// mock openai API
		mux := http.NewServeMux()
		mux.HandleFunc(apiCompletionURL, func(res http.ResponseWriter, req *http.Request) {
			res.Write([]byte(`{
			  "error": {
				"code": "invalid_api_key",
				"message": "Incorrect API key provided: sk-1234**************************************567.",
				"type": "invalid_request_error"
			  }
			}
			`))
		})
		ts := httptest.NewServer(mux)
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

		mocks.AssertReaction(slackClient, ":spiral_note_pad:", message)
		mocks.AssertReaction(slackClient, ":coffee:", message)
		mocks.AssertRemoveReaction(slackClient, ":coffee:", message)
		mocks.AssertError(slackClient, message, "openai error: Incorrect API key provided: sk-1234**************************************567.")

		actual := commands.Run(message)
		assert.True(t, actual)
	})
}
