package custom_commmands

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCustomCommands(t *testing.T) {
	slackClient := mocks.NewSlackClient(t)
	base := bot.BaseCommand{SlackClient: slackClient}

	lock := mocks.LockInternalMessages()
	defer lock.Unlock()

	cfg := &config.Config{}
	cfg.Set("custom_commands.enabled", true)

	commands := bot.Commands{}
	commands.AddCommand(GetCommand(base, cfg))

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

		mocks.AssertSlackMessage(slackClient, message, "No commands define yet. Use `add command 'your alias' 'command to execute'`")

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

		mocks.AssertSlackMessage(slackClient, message, "Added command: `reply 1`. Just use `alias 1` in future.")

		actual := commands.Run(message)

		assert.True(t, actual)
	})

	t.Run("List commands should list new command", func(t *testing.T) {
		message := msg.Message{}
		message.User = "user1"
		message.Text = "list commands"

		mocks.AssertSlackMessage(slackClient, message, "You defined 1 commands:\n - alias 1: `reply 1`")
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

	t.Run("GetRandom 'alias 1' and export", func(t *testing.T) {
		message := msg.Message{}
		message.User = "user1"
		message.Text = "alias 1"

		mocks.AssertSlackMessage(slackClient, message, "executing command: `reply 1`")

		getMessages := mocks.WaitForQueuedMessages(t, 1)

		actual := commands.Run(message)
		assert.True(t, actual)

		messages := getMessages()

		expected := msg.Message{}
		expected.Text = "reply 1"
		expected.User = "user1"

		assert.Equal(t, expected, messages[0])

		// export
		message = msg.Message{}
		message.User = "user1"
		message.Text = "export commands"

		mocks.AssertSlackMessage(slackClient, message, "```- name: \"\"\n  description: \"\"\n  trigger: alias 1\n  category: \"\"\n  commands:\n    - reply 1\n  examples: []\n```")

		actual = commands.Run(message)
		assert.True(t, actual)
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

		mocks.AssertSlackMessage(slackClient, message, "Okay, I deleted command: `alias 1`")

		actual := commands.Run(message)

		assert.True(t, actual)
	})
}
