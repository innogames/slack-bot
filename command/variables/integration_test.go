package variables

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCustomCommands(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	commands := bot.Commands{}
	variablesCommand := GetCommand(base).(command)
	commands.AddCommand(variablesCommand)

	t.Run("Invalid command", func(t *testing.T) {
		message := msg.Message{}
		message.User = "user1"
		message.Text = "notify me not"

		actual := commands.Run(message)
		assert.False(t, actual)
	})

	t.Run("List empty variables", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "list variables"
		message.User = "user1"

		mocks.AssertSlackMessage(slackClient, message, "No variables define yet. Use `add variable 'defaultServer' 'beta'`")

		actual := commands.Run(message)
		assert.True(t, actual)
	})

	t.Run("Add a variable with invalid syntax", func(t *testing.T) {
		message := msg.Message{}
		message.User = "user1"
		message.Text = "add variable name"
		actual := commands.Run(message)
		assert.False(t, actual)
	})

	t.Run("Add valid variable", func(t *testing.T) {
		message := msg.Message{}
		message.User = "user1"
		message.Text = "add variable 'myKey' 'myValue'"

		mocks.AssertSlackMessage(slackClient, message, "Added variable: `myKey` = `myValue`.")

		actual := commands.Run(message)

		assert.True(t, actual)
	})

	t.Run("List commands should list new variable", func(t *testing.T) {
		message := msg.Message{}
		message.User = "user1"
		message.Text = "list variables"

		slackClient.On("SendMessage", message, "You defined 1 variables:\n - myKey: `myValue`").Return("")
		actual := commands.Run(message)

		assert.True(t, actual)
	})

	t.Run("Template with unknown user", func(t *testing.T) {
		function := variablesCommand.GetTemplateFunction()["customVariable"]

		actual := function.(func(string, string) string)("U123", "myKey")
		assert.Equal(t, "_unknown variable: myKey_", actual)
	})

	t.Run("Template with unknown user", func(t *testing.T) {
		function := variablesCommand.GetTemplateFunction()["customVariable"]

		actual := function.(func(string, string) string)("user1", "myKey2")
		assert.Equal(t, "_unknown variable: myKey2_", actual)
	})

	t.Run("Template with known variable", func(t *testing.T) {
		function := variablesCommand.GetTemplateFunction()["customVariable"]

		actual := function.(func(string, string) string)("user1", "myKey")
		assert.Equal(t, "myValue", actual)
	})

	t.Run("Delete variable with invalid syntax", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "delete variable"
		message.User = "user1"

		actual := commands.Run(message)

		assert.False(t, actual)
	})

	t.Run("Delete variable", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "delete variable myKey"
		message.User = "user1"

		slackClient.On("SendMessage", message, "Okay, I deleted variable: `myKey`").Return("")

		actual := commands.Run(message)

		assert.True(t, actual)
	})
}
