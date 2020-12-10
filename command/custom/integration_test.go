package custom

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCustomCommands(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	commands := bot.Commands{}
	commands.AddCommand(GetCommand(base))

	t.Run("Invalid commands", func(t *testing.T) {
		message := msg.Message{}
		message.User = "user1"
		message.Text = "notify me not"

		actual := commands.Run(message)
		assert.False(t, actual)
	})

	t.Run("List empty commands", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "list commands"
		message.User = "user1"
		slackClient.On("SendMessage", message, "No commands define yet. Use `add command 'your alias' 'command to execute'`").Return("")
		actual := commands.Run(message)
		assert.True(t, actual)
	})

	t.Run("Add a command with invalid syntax", func(t *testing.T) {
		message := msg.Message{}
		message.User = "user1"
		message.Text = "add command alias 1 command 2"
		actual := commands.Run(message)
		assert.False(t, actual)
	})

	t.Run("Add valid command", func(t *testing.T) {
		message := msg.Message{}
		message.User = "user1"
		message.Text = "add command 'alias 1' 'reply 1'"

		slackClient.On("SendMessage", message, "Added command: `reply 1`. Just use `alias 1` in future.").Return("")
		actual := commands.Run(message)

		assert.True(t, actual)
	})

	t.Run("List commands should list new command", func(t *testing.T) {
		message := msg.Message{}
		message.User = "user1"
		message.Text = "list commands"

		slackClient.On("SendMessage", message, "You defined 1 commands:\n - alias 1: `reply 1`").Return("")
		actual := commands.Run(message)

		assert.True(t, actual)
	})

	t.Run("GetRandom any command do nothing", func(t *testing.T) {
		message := msg.Message{}
		message.User = "user1"
		message.Text = "any command"
		actual := commands.Run(message)
		assert.False(t, actual)
	})

	t.Run("GetRandom 'alias 1'", func(t *testing.T) {
		message := msg.Message{}
		message.User = "user1"
		message.Text = "alias 1"

		slackClient.On("SendMessage", message, "executing command: `reply 1`").Return("")

		assert.Empty(t, client.InternalMessages)
		actual := commands.Run(message)
		assert.True(t, actual)
		assert.Len(t, client.InternalMessages, 1)

		message = msg.Message{}
		message.Text = "reply 1"
		message.User = "user1"

		handledEvent := <-client.InternalMessages
		assert.Equal(t, handledEvent, message)
	})

	t.Run("Delete command with invalid syntax", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "delete commands 'alias 1'"
		message.User = "user1"

		actual := commands.Run(message)

		assert.False(t, actual)
	})

	t.Run("Delete command", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "delete command 'alias 1'"
		message.User = "user1"

		slackClient.On("SendMessage", message, "Okay, I deleted command: `alias 1`").Return("")

		actual := commands.Run(message)

		assert.True(t, actual)
	})
}
