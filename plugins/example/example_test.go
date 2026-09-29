package example

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExamplePlugin(t *testing.T) {
	slackClient := mocks.NewSlackClient(t)

	t.Run("echo with default prefix", func(t *testing.T) {
		commands, err := setup(bot.NewPluginContext("example", slackClient, config.Config{}))
		require.NoError(t, err)
		assert.Equal(t, 1, commands.Count())

		message := msg.Message{}
		message.Text = "echo hello world"

		mocks.AssertSlackMessage(slackClient, message, "hello world")

		actual := commands.Run(message)
		assert.True(t, actual)

		help := commands.GetHelp()
		assert.Equal(t, "echo <text>", help[0].Command)
	})

	t.Run("custom prefix via config", func(t *testing.T) {
		cfg := config.Config{
			Plugins: map[string]config.PluginConfig{
				"example": {Config: map[string]any{"prefix": "say"}},
			},
		}
		commands, err := setup(bot.NewPluginContext("example", slackClient, cfg))
		require.NoError(t, err)

		message := msg.Message{}
		message.Text = "say hi"

		mocks.AssertSlackMessage(slackClient, message, "hi")

		actual := commands.Run(message)
		assert.True(t, actual)
	})

	t.Run("empty echo is ignored", func(t *testing.T) {
		commands, err := setup(bot.NewPluginContext("example", slackClient, config.Config{}))
		require.NoError(t, err)

		message := msg.Message{}
		message.Text = "echo"

		actual := commands.Run(message)
		assert.True(t, actual)
	})
}
