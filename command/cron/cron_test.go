package cron

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/mocks"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"strings"
	"testing"
)

func TestCron(t *testing.T) {
	slackClient := &mocks.SlackClient{}

	logger := logrus.New()
	crons := []config.Cron{
		{
			Channel:  "#dev",
			Schedule: "0 0 * * *",
			Commands: []string{
				"reply boo",
				"reply bar",
			},
		},
	}
	command := NewCronCommand(slackClient, logger, crons)
	commands := bot.Commands{}
	commands.AddCommand(command)

	t.Run("List crons", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "list crons"
		slackClient.On("SendMessage", message, mock.MatchedBy(func(input string) bool {
			return strings.HasPrefix(input, "*1 crons:*\n - `0 0 * * *`, next in")
		})).Return("")
		actual := commands.Run(message)
		assert.Equal(t, true, actual)
	})

	t.Run("Test help", func(t *testing.T) {
		help := commands.GetHelp()
		assert.NotNil(t, help)
	})
}
