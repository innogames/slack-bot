package openai

import (
	"net/http"
	"testing"
	"time"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/storage"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
)

func TestDalle(t *testing.T) {
	// init memory based storage
	storage.InitStorage("")

	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	t.Run("test http error", func(t *testing.T) {
		ts := startTestServer(
			t,
			apiDalleGenerateImageURL,
			[]testRequest{
				{
					`{"model":"dall-e-3","prompt":"a nice cat","n":1,"size":"1024x1024"}`,
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
		openaiCfg := defaultConfig
		openaiCfg.APIHost = ts.URL
		openaiCfg.APIKey = "0815pass"

		cfg := &config.Config{}
		cfg.Set("openai", openaiCfg)

		defer ts.Close()

		commands := GetCommands(base, cfg)

		message := msg.Message{}
		message.Text = "dalle a nice cat"

		mocks.AssertReaction(slackClient, ":coffee:", message)
		mocks.AssertRemoveReaction(slackClient, ":coffee:", message)
		mocks.AssertError(slackClient, message, "Incorrect API key provided: sk-1234**************************************567.")

		actual := commands.Run(message)
		time.Sleep(time.Millisecond * 100)
		assert.True(t, actual)
	})

	t.Run("test generate image", func(t *testing.T) {
		ts := startTestServer(
			t,
			apiDalleGenerateImageURL,
			[]testRequest{
				{
					`{"model":"dall-e-3","prompt":"a nice cat","n":1,"size":"1024x1024"}`,
					`	{
						  "created": 1700233554,
						  "data": [
							{
							  "url": "https://example.com/image123",
							  "revised_prompt": "revised prompt 1234"
							}
						  ]
						}`,
					http.StatusUnauthorized,
				},
			},
		)
		openaiCfg := defaultConfig
		openaiCfg.APIHost = ts.URL
		openaiCfg.APIKey = "0815pass"

		cfg := &config.Config{}
		cfg.Set("openai", openaiCfg)

		defer ts.Close()

		commands := GetCommands(base, cfg)

		message := msg.Message{}
		message.Text = "dalle a nice cat"

		mocks.AssertReaction(slackClient, ":coffee:", message)
		mocks.AssertRemoveReaction(slackClient, ":coffee:", message)
		mocks.AssertSlackMessage(slackClient, message, " - revised prompt 1234: <https://example.com/image123|open image>\n")

		actual := commands.Run(message)
		time.Sleep(time.Millisecond * 100)
		assert.True(t, actual)
	})
}
