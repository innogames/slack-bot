package admin

import (
	"strings"
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStatsLog(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	cfg := &config.Config{}
	cfg.AdminUsers = []string{
		"UADMIN",
	}

	statsCommand := newStatsCommand(base, cfg)

	command := bot.Commands{}
	command.AddCommand(statsCommand)

	t.Run("invalid command", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "stats"

		actual := command.Run(message)
		assert.False(t, actual)
	})

	t.Run("display bot statsType", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "bot stats"
		message.User = "UADMIN"

		slackClient.On("SendMessage", message, mock.MatchedBy(func(text string) bool {
			return strings.HasPrefix(text, "Here are some current stats:")
		})).Return("")

		actual := command.Run(message)
		assert.True(t, actual)
	})
}
