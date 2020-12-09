package command

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
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

	t.Run("Test help", func(t *testing.T) {
		help := command.GetHelp()
		assert.Equal(t, 1, len(help))
	})
}

func TestMacro(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	client.InternalMessages = make(chan msg.Message, 2)
	cfg := []config.Command{
		{
			Name: "Test",
			Commands: []string{
				"macro 1",
				"macro {{ .text }}",
			},
			Trigger: "start (?P<text>.*)",
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

	t.Run("test util", func(t *testing.T) {
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
}
