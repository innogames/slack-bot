package bot

import (
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsBotMessage(t *testing.T) {
	bot := Bot{}
	bot.auth = &slack.AuthTestResponse{
		UserID: "BOT",
	}

	t.Run("Is random message", func(t *testing.T) {
		event := &slack.MessageEvent{}
		event.Text = "random text"
		actual := bot.shouldHandleMessage(event)
		assert.Equal(t, false, actual)
	})

	t.Run("Is random message to other user", func(t *testing.T) {
		event := &slack.MessageEvent{}
		event.Text = "<@USER2> random test"

		actual := bot.shouldHandleMessage(event)
		assert.Equal(t, false, actual)
	})

	t.Run("Is bot mentioned", func(t *testing.T) {
		event := &slack.MessageEvent{}
		event.User = "U1234"
		event.Text = "<@BOT> random test"

		actual := bot.shouldHandleMessage(event)
		assert.Equal(t, true, actual)
	})

	t.Run("Pass internal events", func(t *testing.T) {
		event := &slack.MessageEvent{}
		event.SubType = TypeInternal
		event.User = "U1234"
		event.Text = "<@BOT> random test"

		actual := bot.shouldHandleMessage(event)
		assert.Equal(t, true, actual)
	})

	t.Run("Is private channel", func(t *testing.T) {
		event := &slack.MessageEvent{
			Msg: slack.Msg{
				Channel: "DRANDOM",
				User:    "U1234",
			},
		}
		event.Text = "random test"
		actual := bot.shouldHandleMessage(event)
		assert.Equal(t, true, actual)
	})

	t.Run("Is random channel", func(t *testing.T) {
		event := &slack.MessageEvent{
			Msg: slack.Msg{
				Channel: "GRANDOM",
				User:    "U1234",
			},
		}
		event.Text = "random test"
		actual := bot.shouldHandleMessage(event)
		assert.Equal(t, false, actual)
	})

	t.Run("Trim", func(t *testing.T) {
		assert.Equal(t, bot.trimMessage(" "), "")
		assert.Equal(t, bot.trimMessage("<@BOT> random ’test’"), "random 'test'")
	})

}

func BenchmarkTrimMessage(b *testing.B) {
	bot := Bot{}
	bot.auth = &slack.AuthTestResponse{}
	bot.auth.User = "botId"

	message := "<@botId> hallo how are `you’?"

	b.Run("trim", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			bot.trimMessage(message)
		}
	})
}

func BenchmarkShouldHandle(b *testing.B) {
	bot := Bot{}
	bot.auth = &slack.AuthTestResponse{}
	bot.auth.User = "botId"

	var result bool
	b.Run("match", func(b *testing.B) {
		event := &slack.MessageEvent{}
		event.Channel = "D123"
		event.User = "U123"

		for i := 0; i < b.N; i++ {
			result = bot.shouldHandleMessage(event)
		}
		assert.True(b, result)
	})
}
