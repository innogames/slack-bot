package command

import (
	"testing"

	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/mocks"
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

		mocks.AssertSlackMessage(slackClient, message, "This 2 are available template functions:\n- makeSlice\n- slice\n")

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Test help", func(t *testing.T) {
		help := command.GetHelp()
		assert.Equal(t, 2, len(help))
	})
}

func TestMacro(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	client.InternalMessages = make(chan msg.Message, 2)
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

		assert.Empty(t, client.InternalMessages)
		actual := command.Run(message)
		assert.True(t, actual)
		assert.NotEmpty(t, client.InternalMessages)

		handledEvent := <-client.InternalMessages
		assert.Equal(t, handledEvent, msg.Message{
			Text: "macro 1",
		})
		handledEvent = <-client.InternalMessages
		assert.Equal(t, handledEvent, msg.Message{
			Text: "macro test",
		})
		assert.Empty(t, client.InternalMessages)
	})

	t.Run("test error", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "test backend"

		mocks.AssertError(slackClient, message, "template: reply {{project}}:1: function \"project\" not defined")

		actual := command.Run(message)
		assert.True(t, actual)
	})
}
