package example

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
)

func TestExamplePlugin(t *testing.T) {
	slackClient := mocks.NewSlackClient(t)

	t.Run("echo with default prefix", func(t *testing.T) {
		cfg := config.Config{}
		commands := getCommands(slackClient, cfg)
		assert.Equal(t, 1, commands.Count())

		message := msg.Message{}
		message.Text = "echo hello world"

		mocks.AssertSlackMessage(slackClient, message, "hello world")

		actual := commands.Run(message)
		assert.True(t, actual)
	})

	t.Run("custom prefix via config", func(t *testing.T) {
		cfg := config.Config{}
		cfg.Set("plugins§example", map[string]any{"prefix": "say"})
		commands := getCommands(slackClient, cfg)

		message := msg.Message{}
		message.Text = "say hi"

		mocks.AssertSlackMessage(slackClient, message, "hi")

		actual := commands.Run(message)
		assert.True(t, actual)
	})

	t.Run("empty echo is ignored", func(t *testing.T) {
		cfg := config.Config{}
		commands := getCommands(slackClient, cfg)

		message := msg.Message{}
		message.Text = "echo"

		actual := commands.Run(message)
		assert.True(t, actual)
	})
}
