package command

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
)

func TestInvalidMacro(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	client.InternalMessages = make(chan msg.Message, 2)
	cfg := []config.Command{
		{
			Name: "Test",
			Commands: []string{
				"macro 1",
			},
			Category: "Test",
			Trigger:  "start test",
		},
	}

	command := bot.Commands{}
	command.AddCommand(NewCommands(base, cfg))

	t.Run("invalid command", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "start foo"

		actual := command.Run(message)
		assert.False(t, actual)
	})

	t.Run("list template functions", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "list template functions"

		mocks.AssertSlackMessage(slackClient, message, "*There are 2 available template functions:*\n• makeSlice(...interface {}) []interface {}\n• slice(string, int, int) string\n")

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Test help", func(t *testing.T) {
		help := command.GetHelp()
		assert.Equal(t, 2, len(help))
	})
}

func TestDefinedCommands(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	lock := mocks.LockInternalMessages()
	defer lock.Unlock()

	cfg := []config.Command{
		{
			Name: "Test 1",
			Commands: []string{
				"macro 1",
				"macro {{ .text }}",
			},
			Trigger: "start (?P<text>.*)",
		},
		{
			Name: "Test 2",
			Commands: []string{
				"reply {{project}}", // should be {{.project}} to output the var
			},
			Trigger: "test (?P<project>(backend|mobile|frontend))",
		},
	}

	command := bot.Commands{}
	command.AddCommand(NewCommands(base, cfg))

	t.Run("invalid macro", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "commandHelp quatsch"

		actual := command.Run(message)
		assert.False(t, actual)
	})

	t.Run("invalid macro with regexp", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "test 122"

		actual := command.Run(message)
		assert.False(t, actual)
	})

	t.Run("test commands", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "start test"

		getMessages := mocks.WaitForQueuedMessages(t, 2)

		// run the command -> it's blocking until all sub-commands are handled
		assert.Empty(t, client.InternalMessages)
		actual := command.Run(message)
		assert.True(t, actual)

		actualMessages := getMessages()

		assert.Equal(t, "macro 1", actualMessages[0].Text)
		assert.Equal(t, "macro test", actualMessages[1].Text)
	})

	t.Run("test error", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "test backend"

		mocks.AssertError(slackClient, message, "template: reply {{project}}:1: function \"project\" not defined")

		actual := command.Run(message)
		assert.True(t, actual)
	})
}
