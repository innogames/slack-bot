package admin

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/mocks"
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
		message := msg.Message{}
		message.Text = "stats"

		actual := command.Run(message)
		assert.Equal(t, false, actual)
	})

	t.Run("display bot statsType", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "bot stats"
		message.User = "UADMIN"

		slackClient.On("SendMessage", message, mock.MatchedBy(func(text string) bool {
			return strings.HasPrefix(text, "Here are some current stats:")
		})).Return("")

		actual := command.Run(message)
		assert.Equal(t, true, actual)
	})
}
