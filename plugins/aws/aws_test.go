package aws

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
)

func TestGetCommands(t *testing.T) {
	slackClient := mocks.NewSlackClient(t)

	awsConfig := map[string]any{
		"enabled": true,
		"cloud_front": []map[string]any{
			{"id": "id", "name": "name"},
		},
	}

	t.Run("AWS is not active", func(t *testing.T) {
		cfg := config.Config{}
		commands := getCommands(slackClient, cfg)
		assert.Equal(t, 0, commands.Count())
	})

	t.Run("AWS is disabled via plugins section", func(t *testing.T) {
		cfg := config.Config{}
		cfg.Set("plugins§aws", map[string]any{"enabled": false})
		commands := getCommands(slackClient, cfg)
		assert.Equal(t, 0, commands.Count())
	})

	t.Run("AWS is active via plugins section", func(t *testing.T) {
		cfg := config.Config{}
		cfg.Set("plugins§aws", awsConfig)
		commands := getCommands(slackClient, cfg)
		assert.Equal(t, 2, commands.Count())

		// test help
		help := commands.GetHelp()
		assert.Len(t, help, 2)

		// list the CF
		message := msg.Message{}
		message.Text = "aws cf list"
		mocks.AssertSlackBlocks(t, slackClient, message, `[{"type":"section","text":{"type":"mrkdwn","text":"\"id\": name\n"}}]`)

		actual := commands.Run(message)
		assert.True(t, actual)
	})

	t.Run("AWS is active via legacy aws key", func(t *testing.T) {
		cfg := config.Config{}
		cfg.Set("aws", awsConfig)
		commands := getCommands(slackClient, cfg)
		assert.Equal(t, 2, commands.Count())
	})
}
