package bot

import (
	"fmt"
	"testing"

	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/config"
	"github.com/nlopes/slack"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestBot(t *testing.T) {
	cfg := config.Config{}

	slackClient := &client.Slack{}
	slackClient.RTM = *slackClient.NewRTM()

	logger, _ := test.NewNullLogger()

	commands := &Commands{}
	commands.AddCommand(testCommand2{})

	bot := NewBot(cfg, slackClient, logger, commands)
	bot.auth = &slack.AuthTestResponse{
		UserID: "BOT",
	}
	bot.allowedUsers = map[string]string{
		"U123": "User123",
	}

	t.Run("handle empty message", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = ""
		event.Channel = "C123"
		bot.handleMessage(event)
	})

	t.Run("handle unauthenticated message", func(t *testing.T) {
		event := slack.MessageEvent{}

		event.Text = "test"
		event.User = "U888"
		event.Channel = "C123"
		bot.handleMessage(event)
	})

	t.Run("handle valid message", func(t *testing.T) {
		event := slack.MessageEvent{}

		event.Text = "test"
		event.User = "U123"
		event.Channel = "C123"
		bot.handleMessage(event)
		fmt.Println(bot.slackClient.RTM.IncomingEvents)
	})
}

func TestIsBotMessage(t *testing.T) {
	bot := bot{}
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
	bot := bot{}
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
	bot := bot{}
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

type testCommand2 struct {
}

func (c testCommand2) GetMatcher() matcher.Matcher {
	return matcher.NewTextMatcher("test", func(match matcher.Result, event slack.MessageEvent) {
	})
}
