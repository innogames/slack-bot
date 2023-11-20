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
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDalle(t *testing.T) {
	// init memory based storage
	storage.InitStorage("")

	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	t.Run("test http error", func(t *testing.T) {
		openaiCfg, ts := startTestServer(
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
		openaiCfg, ts := startTestServer(
			t,
			apiDalleGenerateImageURL,
			[]testRequest{
				{
					`{"model":"dall-e-3","prompt":"a nice cat","n":1,"size":"1024x1024"}`,
					`	{
						  "created": 1700233554,
						  "data": [
							{
							  "url": "{test_server}/v1/images/generations",
							  "revised_prompt": "revised prompt 1234"
							}
						  ]
						}`,
					http.StatusUnauthorized,
				},
				{
					``,
					`just something`,
					http.StatusOK,
				},
			},
		)

		cfg := &config.Config{}
		cfg.Set("openai", openaiCfg)

		defer ts.Close()

		commands := GetCommands(base, cfg)

		message := msg.Message{}
		message.Text = "dalle a nice cat"

		mocks.AssertReaction(slackClient, ":coffee:", message)
		mocks.AssertRemoveReaction(slackClient, ":coffee:", message)

		slackClient.On(
			"UploadFile",
			mock.MatchedBy(func(params slack.FileUploadParameters) bool {
				assert.Equal(t, "dalle.png", params.Filename)
				return true
			}),
		).Return(nil, nil).Once()

		actual := commands.Run(message)
		time.Sleep(time.Millisecond * 100)
		assert.True(t, actual)
	})
}
