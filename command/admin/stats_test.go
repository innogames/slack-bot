package admin

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/mocks"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"strings"
	"testing"
)

func TestStatsLog(t *testing.T) {
	slackClient := mocks.SlackClient{}
	cfg := config.Config{}
	cfg.AdminUsers = []string{
		"UADMIN",
	}

	statsCommand := NewStatsCommand(&slackClient, cfg)

	command := bot.Commands{}
	command.AddCommand(statsCommand)

	t.Run("invalid command", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "stats"

		actual := command.Run(event)
		assert.Equal(t, false, actual)
	})

	t.Run("display bot statsType", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "bot stats"
		event.User = "UADMIN"

		slackClient.On("Reply", event, mock.MatchedBy(func(text string) bool {
			return strings.HasPrefix(text, "Here are some current stats:")
		}))

		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})
}
