package command

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHelp(t *testing.T) {
	cfg := config.Config{}
	cfg.Jenkins.Host = "bitbucket.example.com"
	slackClient := &mocks.SlackClient{}
	slackClient.On("CanHandleInteractions").Return(true)

	base := bot.BaseCommand{SlackClient: slackClient}

	commands := GetCommands(slackClient, cfg)

	help := bot.Commands{}
	help.AddCommand(NewHelpCommand(base, commands))

	t.Run("invalid command", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "no help"

		actual := help.Run(message)
		assert.False(t, actual)
	})

	t.Run("list all commands", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "help"

		mocks.AssertReaction(slackClient, "💡", message)
		slackClient.On("SendEphemeralMessage", message, mock.AnythingOfType("string"))
		actual := help.Run(message)
		assert.True(t, actual)
	})

	t.Run("help for specific command", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "help reply"

		slackClient.On("SendEphemeralMessage", message, "*reply <text>*:\njust reply the given message\n*Some examples:*\n - reply Hello, how are you?\n")
		actual := help.Run(message)
		assert.True(t, actual)
	})

	t.Run("help for invalid command", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "help sdsadasdasd"

		slackClient.On("SendEphemeralMessage", message, "Invalid command: `sdsadasdasd`")

		actual := help.Run(message)
		assert.True(t, actual)
	})
}
