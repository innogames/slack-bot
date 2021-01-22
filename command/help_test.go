package command

import (
	"testing"

	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/mocks"
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

		slackClient.On("SendMessage", message, mock.AnythingOfType("string")).Return("")
		actual := help.Run(message)
		assert.True(t, actual)
	})

	t.Run("help for specific command", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "help reply"

		slackClient.On("SendMessage", message, "*reply command*:\njust reply the given message\n*Some examples:*\n - reply Hello, how are you?\n").Return("")
		actual := help.Run(message)
		assert.True(t, actual)
	})

	t.Run("help for invalid command", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "help sdsadasdasd"

		mocks.AssertSlackMessage(slackClient, message, "Invalid command: `sdsadasdasd`")

		actual := help.Run(message)
		assert.True(t, actual)
	})
}
