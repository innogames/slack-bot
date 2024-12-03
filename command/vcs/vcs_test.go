package vcs

import (
	"testing"
	"time"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
)

// compile time check that the interface matches
var _ bot.Runnable = &vcsCommand{}

func TestVCS(t *testing.T) {
	slackClient := mocks.NewSlackClient(t)
	base := bot.BaseCommand{SlackClient: slackClient}

	t.Run("Test disabled", func(t *testing.T) {
		cfg := &config.Config{}
		commands := GetCommands(base, cfg)
		assert.Equal(t, 0, commands.Count())
	})

	t.Run("list branches", func(t *testing.T) {
		cfg := &config.Config{}
		cfg.BranchLookup = config.VCS{
			Type:           "git",
			UpdateInterval: time.Second,
		}

		message := msg.Message{}
		message.Text = "list branches"

		mocks.AssertSlackMessage(slackClient, message, "Found 0 branches:\n")

		commands := GetCommands(base, cfg)
		assert.Equal(t, 1, commands.Count())
		assert.Len(t, commands.GetHelp(), 1)

		actual := commands.Run(message)
		time.Sleep(100 * time.Millisecond)
		assert.True(t, actual)
	})
}
