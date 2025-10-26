package openai

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/util"
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

		// Reactions happen asynchronously and timing can vary, especially on slower CI (Windows)
		// Use Maybe() to make the test focus on chunking behavior, not reaction timing
		slackClient.On("AddReaction", mock.MatchedBy(func(actualReaction util.Reaction) bool {
			return util.Reaction(":bulb:").ToSlackReaction() == actualReaction.ToSlackReaction()
		}), ref).Maybe()
		slackClient.On("AddReaction", mock.MatchedBy(func(actualReaction util.Reaction) bool {
			return util.Reaction(":speech_balloon:").ToSlackReaction() == actualReaction.ToSlackReaction()
		}), ref).Maybe()
		slackClient.On("RemoveReaction", util.Reaction(":bulb:"), ref).Maybe()
		slackClient.On("RemoveReaction", util.Reaction(":speech_balloon:"), ref).Maybe()

		// When chunking occurs, multiple SendMessage calls are made
		// Initial message with placeholder (from command.go)
		slackClient.On("SendMessage", ref, ":bulb: thinking...", mock.Anything).Return("msg1").Once()
		// Additional messages for new chunks (from chunker.go) - allow multiple calls
		slackClient.On("SendMessage", ref, ":bulb: thinking...", mock.Anything, mock.Anything).Return("msg2").Maybe()
		// Final updates for all chunks - expecting any string content with 2 MsgOptions (MsgOptionUpdate and MsgOptionTS)
		slackClient.On("SendMessage", ref, mock.MatchedBy(func(s string) bool { return len(s) > 0 }), mock.Anything, mock.Anything).Return("").Maybe()

		actual := commands.Run(message)
		// Give the goroutine time to start and complete reactions
		// This is especially important on Windows CI which may be slower
		time.Sleep(time.Millisecond * 100)
		queue.WaitTillHavingNoQueuedMessage()
		assert.True(t, actual)
	})
}
