package pool

import (
	"testing"
	"time"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
)

func TestPools(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	t.Run("Pools are not active", func(t *testing.T) {
		cfg := &config.Pool{}
		commands := GetCommands(cfg, base)
		assert.Equal(t, 0, commands.Count())
	})

	t.Run("Full test", func(t *testing.T) {
		cfg := &config.Pool{
			LockDuration: time.Minute,
			NotifyExpire: time.Minute,
			Resources: []*config.Resource{
				{
					Name: "server1",
				},
				{
					Name: "server2",
				},
			},
		}
		commands := GetCommands(cfg, base)
		assert.Equal(t, 1, commands.Count())

		runCommand := func(message msg.Message) msg.Message {
			actual := commands.Run(message)
			assert.True(t, actual)

			return message
		}

		// list
		message := msg.Message{}
		message.Text = "pool list"
		mocks.AssertSlackMessage(slackClient, message, "*Available:*\n`server1`, `server2`\n\n*Used/Locked:*")
		message = runCommand(message)

		// lock
		message = msg.Message{}
		message.Text = "pool lock server1"
		mocks.AssertSlackMessageRegexp(slackClient, message, "^`server1` is locked for you until")
		message = runCommand(message)

		// extend
		message = msg.Message{}
		message.Text = "pool extend server1 1h"
		mocks.AssertSlackMessageRegexp(slackClient, message, "^`server1` got extended until")
		runCommand(message)

		// pool locks
		message = msg.Message{}
		message.Text = "pool locks"
		mocks.AssertSlackMessageRegexp(slackClient, message, "^ \\*Your locks:\\*\n\n`server1` until")
		runCommand(message)

		// unlock
		message = msg.Message{}
		message.Text = "pool unlock server1"
		mocks.AssertSlackMessage(slackClient, message, "`server1` is free again")
		runCommand(message)

		help := commands.GetHelp()
		assert.Equal(t, 6, len(help))
	})
}
