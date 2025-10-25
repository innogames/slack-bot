package openai

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/command/queue"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMessageChunking(t *testing.T) {
	t.Run("Test long response with message chunking", func(t *testing.T) {
		// Create a very long response (>7000 chars) to test chunking
		longPart1 := strings.Repeat("X", 3600) // Exceeds 3500 char limit
		longPart2 := strings.Repeat("Y", 3600) // Exceeds 3500 char limit again

		openaiCfg, ts := startTestServer(
			t,
			apiCompletionURL,
			[]testRequest{
				{
					`{"model":"gpt-4o","messages":[{"role":"system","content":"You are a helpful Slack bot. By default, keep your answer short and truthful"},{"role":"user","content":"give me a long response"}],"stream":true}`,
					fmt.Sprintf(`data: {"id":"chatcmpl-6tuxebSPdmd2IJpb8GrZXHiYXON6r","object":"chat.completion.chunk","created":1678785018,"model":"gpt-4o-0301","choices":[{"delta":{"role":"assistant"},"index":0,"finish_reason":null}]}

data: {"id":"chatcmpl-6tuxebSPdmd2IJpb8GrZXHiYXON6r","object":"chat.completion.chunk","created":1678785018,"model":"gpt-4o-0301","choices":[{"delta":{"content":"%s"},"index":0,"finish_reason":null}]}

data: {"id":"chatcmpl-6tuxebSPdmd2IJpb8GrZXHiYXON6r","object":"chat.completion.chunk","created":1678785018,"model":"gpt-4o-0301","choices":[{"delta":{"content":"%s"},"index":0,"finish_reason":null}]}

data: {"id":"chatcmpl-6tuxebSPdmd2IJpb8GrZXHiYXON6r","object":"chat.completion.chunk","created":1678785018,"model":"gpt-4o-0301","choices":[{"delta":{},"index":0,"finish_reason":"stop"}]}

data: [DONE]`, longPart1, longPart2),
					http.StatusOK,
				},
			},
		)
		defer ts.Close()

		slackClient := mocks.NewSlackClient(t)
		base := bot.BaseCommand{SlackClient: slackClient}

		openaiCfg.LogTexts = true
		cfg := &config.Config{}
		cfg.Set("openai", openaiCfg)

		commands := GetCommands(base, cfg)

		message := msg.Message{}
		message.Text = "openai give me a long response"
		message.Channel = "testchan"
		message.Timestamp = "1234"
		ref := message.MessageRef

		mocks.AssertReaction(slackClient, ":bulb:", ref)
		mocks.AssertReaction(slackClient, ":speech_balloon:", ref)
		mocks.AssertRemoveReaction(slackClient, ":bulb:", ref)
		mocks.AssertRemoveReaction(slackClient, ":speech_balloon:", ref)

		// When chunking occurs, multiple SendMessage calls are made
		// First message with initial placeholder
		slackClient.On("SendMessage", ref, ":bulb: thinking...", mock.Anything).Return("msg1").Once()
		// Second message for second chunk with placeholder
		slackClient.On("SendMessage", ref, ":bulb: thinking...", mock.Anything, mock.Anything).Return("msg2").Once()
		// Third message for third chunk with placeholder
		slackClient.On("SendMessage", ref, ":bulb: thinking...", mock.Anything, mock.Anything).Return("msg3").Once()
		// Final updates for all chunks - expecting any string content with 2 MsgOptions (MsgOptionUpdate and MsgOptionTS)
		slackClient.On("SendMessage", ref, mock.MatchedBy(func(s string) bool { return len(s) > 0 }), mock.Anything, mock.Anything).Return("").Maybe()

		actual := commands.Run(message)
		queue.WaitTillHavingNoQueuedMessage()
		assert.True(t, actual)
	})
}
