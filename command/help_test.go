package command

import (
	"testing"

	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/storage"
	"github.com/innogames/slack-bot/mocks"
	"github.com/nlopes/slack"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHelp(t *testing.T) {
	cfg := config.Config{}
	cfg.Jenkins.Host = "bitbucket.example.com"
	logger := logrus.New()
	slackClient := &mocks.SlackClient{}

	storage.MockStorage()

	commands := GetCommands(slackClient, cfg, logger)

	help := bot.Commands{}
	help.AddCommand(NewHelpCommand(slackClient, commands))

	t.Run("invalid command", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "no help"

		actual := help.Run(event)
		assert.Equal(t, false, actual)
	})

	t.Run("list all commands", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "help"

		slackClient.On("Reply", event, mock.AnythingOfType("string"))
		actual := help.Run(event)
		assert.Equal(t, true, actual)
	})

	t.Run("help for specific command", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "help reply"

		slackClient.On("Reply", event, "*reply command*:\njust reply the given message\n*Some examples:*\n - reply Hello, how are you?\n")
		actual := help.Run(event)
		assert.Equal(t, true, actual)
	})

	t.Run("help for invalid command", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "help sdsadasdasd"

		slackClient.On("Reply", event, "Invalid command: `sdsadasdasd`")
		actual := help.Run(event)
		assert.Equal(t, true, actual)
	})
}
