package aws

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetup(t *testing.T) {
	slackClient := mocks.NewSlackClient(t)

	cfg := config.Config{
		Plugins: map[string]config.PluginConfig{
			"aws": {
				Config: map[string]any{
					"cloud_front": []any{
						map[string]any{"id": "id", "name": "name"},
					},
				},
			},
		},
	}

	commands, err := setup(bot.NewPluginContext("aws", slackClient, cfg))
	require.NoError(t, err)
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
}
