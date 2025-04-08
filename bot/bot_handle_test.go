package bot

import (
	"testing"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestIsBotMessage(t *testing.T) {
	bot := Bot{}
	bot.auth = &slack.AuthTestResponse{
		UserID: "BOT",
	}

	t.Run("Is random message", func(t *testing.T) {
		event := &slack.MessageEvent{}
		event.Text = "random text"
		actual := bot.canHandleMessage(event)
		assert.False(t, actual)
	})

	t.Run("Is random message to other user", func(t *testing.T) {
		event := &slack.MessageEvent{}
		event.Text = "<@USER2> random test"

		actual := bot.canHandleMessage(event)
		assert.False(t, actual)
	})

	t.Run("Is Bot mentioned", func(t *testing.T) {
		event := &slack.MessageEvent{}
		event.User = "U1234"
		event.Text = "<@BOT> random test"

		actual := bot.canHandleMessage(event)
		assert.True(t, actual)
	})

	t.Run("Pass internal events", func(t *testing.T) {
		event := &slack.MessageEvent{}
		event.User = "U1234"
		event.Text = "<@BOT> random test"

		actual := bot.canHandleMessage(event)
		assert.True(t, actual)
	})

	t.Run("Disables authentication", func(t *testing.T) {
		bot.config.NoAuthentication = true
		userID := "U1233"

		actual := bot.isUserActionAllowed(userID)
		assert.True(t, actual)
	})

	t.Run("Enabled authentication authentication", func(t *testing.T) {
		bot.config.NoAuthentication = false
		userID := "U1233"

		actual := bot.isUserActionAllowed(userID)
		assert.False(t, actual)
	})

	t.Run("Is private channel", func(t *testing.T) {
		event := &slack.MessageEvent{
			Msg: slack.Msg{
				Channel: "DRANDOM",
				User:    "U1234",
			},
		}
		event.Text = "random test"
		actual := bot.canHandleMessage(event)
		assert.True(t, actual)
	})

	t.Run("Is random channel", func(t *testing.T) {
		event := &slack.MessageEvent{
			Msg: slack.Msg{
				Channel: "GRANDOM",
				User:    "U1234",
			},
		}
		event.Text = "random test"
		actual := bot.canHandleMessage(event)
		assert.False(t, actual)
	})

	t.Run("Trim + Clean", func(t *testing.T) {
		assert.Empty(t, bot.cleanMessage(" ", true))
		assert.Equal(t, "random 'test'", bot.cleanMessage("*<@BOT> random ’test’*", true))
		assert.Equal(t, "random 'test'", bot.cleanMessage("<@BOT> random ’test’", true))
		assert.Equal(t, "random Ananas Banane", bot.cleanMessage("<@BOT> random Ananas Banane", true))
		assert.Equal(t, `add button "test" "reply it works"`, bot.cleanMessage("add button “test” “reply it works”", true))
		assert.Equal(t, "<https://test.com|TEST> <https://example.com|example>", bot.cleanMessage("<https://test.com|TEST> <https://example.com|example>", false))
		assert.Equal(t, "TEST example", bot.cleanMessage("<https://test.com|TEST> <https://example.com|example>", true))
	})
}

func BenchmarkTrimMessage(b *testing.B) {
	bot := Bot{}
	bot.auth = &slack.AuthTestResponse{}
	bot.auth.User = "botId"

	message := "<@botId> hallo how are `you’?"

	b.Run("trim", func(b *testing.B) {
		for range b.N {
			bot.cleanMessage(message, false)
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

		for range b.N {
			result = bot.canHandleMessage(event)
		}
		assert.True(b, result)
	})
}
