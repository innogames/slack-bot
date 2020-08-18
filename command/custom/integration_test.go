package custom

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/mocks"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCustomCommands(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	commands := bot.Commands{}
	commands.AddCommand(GetCommand(slackClient))

	t.Run("Invalid commands", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.User = "user1"
		event.Text = "notify me not"

		actual := commands.Run(event)
		assert.Equal(t, false, actual)
	})

	t.Run("List empty commands", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "list commands"
		event.User = "user1"
		slackClient.On("Reply", event, "No commands define yet. Use `add command 'your alias' 'command to execute'`").Return("")
		actual := commands.Run(event)
		assert.Equal(t, true, actual)
	})

	t.Run("Add a command with invalid syntax", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.User = "user1"
		event.Text = "add command alias 1 command 2"
		actual := commands.Run(event)
		assert.Equal(t, false, actual)
	})

	t.Run("Add valid command", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.User = "user1"
		event.Text = "add command 'alias 1' 'reply 1'"

		slackClient.On("Reply", event, "Added command: `reply 1`. Just use `alias 1` in future.").Return("")
		actual := commands.Run(event)

		assert.Equal(t, true, actual)
	})

	t.Run("List commands should list new command", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.User = "user1"
		event.Text = "list commands"

		slackClient.On("Reply", event, "You defined 1 commands:\n - alias 1: `reply 1`")
		actual := commands.Run(event)

		assert.Equal(t, true, actual)
	})

	t.Run("GetRandom any command do nothing", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.User = "user1"
		event.Text = "any command"
		actual := commands.Run(event)
		assert.Equal(t, false, actual)
	})

	t.Run("GetRandom 'alias 1'", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.User = "user1"
		event.Text = "alias 1"

		slackClient.On("Reply", event, "executing command: `reply 1`")

		assert.Empty(t, client.InternalMessages)
		actual := commands.Run(event)
		assert.Equal(t, true, actual)
		assert.Len(t, client.InternalMessages, 1)

		event = slack.MessageEvent{}
		event.Text = "reply 1"
		event.User = "user1"

		handledEvent := <-client.InternalMessages
		assert.Equal(t, handledEvent, event)
	})

	t.Run("Delete command with invalid syntax", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "delete commands 'alias 1'"
		event.User = "user1"

		actual := commands.Run(event)

		assert.Equal(t, false, actual)
	})

	t.Run("Delete command", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "delete command 'alias 1'"
		event.User = "user1"

		slackClient.On("Reply", event, "Okay, I deleted command: `alias 1`")

		actual := commands.Run(event)

		assert.Equal(t, true, actual)
	})
}
