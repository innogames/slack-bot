package pool

import (
	"github.com/innogames/slack-bot/v2/bot/config"
	"testing"
	"time"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNumberGuesser(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	t.Run("Pools are not active", func(t *testing.T) {
		cfg := &config.Pool{}
		commands := GetCommands(cfg, base)
		assert.Equal(t, 0, commands.Count())
	})

	t.Run("Pools are active", func(t *testing.T) {
		cfg := &config.Pool{
			LockDuration: time.Minute,
			NotifyExpire: time.Minute,
			Resources: []*config.Resource{
				{
					Name: "fooo resource",
				},
			},
		}
		commands := GetCommands(cfg, base)
		assert.Equal(t, 1, commands.Count())
	})

}
