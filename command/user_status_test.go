package command

import (
	"fmt"
	"testing"
	"time"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/command/queue"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestUserStatus(t *testing.T) {
	slackClient := &mocks.SlackClient{}

	base := bot.BaseCommand{SlackClient: slackClient}
	command := newUserStatusCommand(base)

	commands := bot.Commands{}
	commands.AddCommand(command)

	t.Run("Invalid command", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "notify for something"

		actual := commands.Run(message)
		assert.False(t, actual)
	})

	t.Run("Check with error", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "notify user <@U123456> active"

		err := fmt.Errorf("some slack error")
		slackClient.On("GetUserPresence", "U123456").Once().Return(nil, err)

		mocks.AssertReaction(slackClient, "⌛", message)
		mocks.AssertRemoveReaction(slackClient, "⌛", message)
		mocks.AssertError(slackClient, message, err)

		actual := commands.Run(message)
		assert.True(t, actual)
		queue.WaitTillHavingNoQueuedMessage()
	})

	t.Run("Check user getting active", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "notify user <@U123456> active"

		command.checkInterval = time.Millisecond * 1
		presenceAway := &slack.UserPresence{
			Presence: "away",
		}
		presenceActive := &slack.UserPresence{
			Presence: "active",
		}
		slackClient.On("GetUserPresence", "U123456").Once().Return(presenceAway, nil)
		slackClient.On("GetUserPresence", "U123456").Once().Return(presenceActive, nil)

		mocks.AssertReaction(slackClient, "⌛", message)
		mocks.AssertRemoveReaction(slackClient, "⌛", message)
		mocks.AssertSlackMessage(slackClient, message, "User <@U123456> is active now!")

		actual := commands.Run(message)
		assert.True(t, actual)
		queue.WaitTillHavingNoQueuedMessage()
	})
}
