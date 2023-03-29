package admin

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/stats"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
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

		stats.Increase("handled_command_help", 1)
		mocks.AssertSlackMessageRegexp(slackClient, message, "(?s)^Here are some current stats:.*command help: 1.*")

		actual := command.Run(message)
		assert.True(t, actual)
	})
}
