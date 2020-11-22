package command

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/mocks"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInvalidMacro(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	logger := logrus.New()

	client.InternalMessages = make(chan msg.Message, 2)
	cfg := []config.Macro{
		{
			Name: "Test",
			Commands: []string{
				"macro 1",
			},
			Trigger: "start test",
		},
	}

	command := bot.Commands{}
	command.AddCommand(NewMacroCommand(slackClient, cfg, logger))

	t.Run("invalid command", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "start foo"

		actual := command.Run(event)
		assert.Equal(t, false, actual)
	})
}

func TestMacro(t *testing.T) {
	logger := logrus.New()

	slackClient := &mocks.SlackClient{}
	client.InternalMessages = make(chan msg.Message, 2)
	cfg := []config.Macro{
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
	command.AddCommand(NewMacroCommand(slackClient, cfg, logger))

	t.Run("invalid macro", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "help quatsch"

		actual := command.Run(event)
		assert.Equal(t, false, actual)
	})

	t.Run("test util", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "start test"

		assert.Empty(t, client.InternalMessages)
		actual := command.Run(event)
		assert.Equal(t, true, actual)
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
