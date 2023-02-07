package aws

import (
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetCommands(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	t.Run("AWS is not active", func(t *testing.T) {
		cfg := config.Aws{}
		commands := GetCommands(cfg, base)
		assert.Equal(t, 0, commands.Count())
	})

	t.Run("AWS is active", func(t *testing.T) {
		cfg := config.Aws{
			Enabled: true,
			CloudFront: []config.AwsCfDistribution{
				{ID: "id", Name: "name"},
			},
		}
		commands := GetCommands(cfg, base)
		assert.Equal(t, 2, commands.Count())
	})
}
