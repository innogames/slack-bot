package variables

import (
	"testing"

	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/mocks"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestCustomCommands(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	commands := bot.Commands{}
	variablesCommand := GetCommand(slackClient).(command)
	commands.AddCommand(variablesCommand)

	t.Run("Invalid command", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.User = "user1"
		event.Text = "notify me not"

		actual := commands.Run(event)
		assert.Equal(t, false, actual)
	})

	t.Run("List empty variables", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "list variables"
		event.User = "user1"
		slackClient.On("Reply", event, "No variables define yet. Use `add variable 'defaultServer' 'beta'`").Return("")
		actual := commands.Run(event)
		assert.Equal(t, true, actual)
	})

	t.Run("Add a variable with invalid syntax", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.User = "user1"
		event.Text = "add variable name"
		actual := commands.Run(event)
		assert.Equal(t, false, actual)
	})

	t.Run("Add valid variable", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.User = "user1"
		event.Text = "add variable 'myKey' 'myValue'"

		slackClient.On("Reply", event, "Added variable: `myKey` = `myValue`.").Return("")
		actual := commands.Run(event)

		assert.Equal(t, true, actual)
	})

	t.Run("List commands should list new variable", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.User = "user1"
		event.Text = "list variables"

		slackClient.On("Reply", event, "You defined 1 variables:\n - myKey: `myValue`")
		actual := commands.Run(event)

		assert.Equal(t, true, actual)
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
		event := slack.MessageEvent{}
		event.Text = "delete variable"
		event.User = "user1"

		actual := commands.Run(event)

		assert.Equal(t, false, actual)
	})

	t.Run("Delete variable", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "delete variable myKey"
		event.User = "user1"

		slackClient.On("Reply", event, "Okay, I deleted variable: `myKey`")

		actual := commands.Run(event)

		assert.Equal(t, true, actual)
	})

}
