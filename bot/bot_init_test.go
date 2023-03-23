package bot

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestBot(t *testing.T) {
	cfg := config.Config{}
	cfg.Slack.Token = "xoxb-12345"
	cfg.AdminUsers = config.UserList{
		"admin123",
	}

	slackClient, err := client.GetSlackClient(cfg.Slack)
	assert.Nil(t, err)

	commands := &Commands{}
	commands.AddCommand(testCommand2{})

	commandNames := commands.GetCommandNames()
	expectedNames := []string{"bot.testCommand2"}
	assert.Equal(t, expectedNames, commandNames)

	bot := NewBot(cfg, slackClient, commands)
	bot.auth = &slack.AuthTestResponse{
		UserID: "BOT",
	}
	bot.allowedUsers = map[string]string{
		"U123": "User123",
	}

	t.Run("handle empty message", func(t *testing.T) {
		message := msg.Message{}
		message.Text = ""
		message.Channel = "C123"
		bot.ProcessMessage(message, true)
	})

	t.Run("handle unauthenticated message", func(t *testing.T) {
		message := msg.Message{}

		message.Text = "test"
		message.User = "U888"
		message.Channel = "C123"
		bot.ProcessMessage(message, true)
	})

	t.Run("handle valid message", func(t *testing.T) {
		message := msg.Message{}

		message.Text = "test"
		message.User = "U123"
		message.Channel = "C123"
		bot.ProcessMessage(message, true)
	})

	t.Run("Init with invalid token", func(t *testing.T) {
		bot.config.Slack.Token = "invalid"

		err := bot.Init()
		assert.EqualError(t, err, "auth error: invalid_auth")
	})

	/*
		todo race
			t.Run("Init with invalid timezone", func(t *testing.T) {
				bot.config.Timezone = "foo/bar"

				err := bot.Init()
				assert.EqualError(t, err, "unknown time zone foo/bar")
			})
	*/

	t.Run("Load channels without token", func(t *testing.T) {
		channels, err := bot.loadChannels()
		assert.Len(t, channels, 0)
		assert.Contains(t, err.Error(), "invalid_auth")
	})
}

func TestInitLogger(t *testing.T) {
	_ = t

	cfg := config.Config{}
	cfg.Logger.Level = "debug"
	InitLogger(cfg.Logger)
}

type testCommand2 struct{}

func (c testCommand2) GetMatcher() matcher.Matcher {
	return matcher.NewTextMatcher("test", func(match matcher.Result, message msg.Message) {
	})
}
