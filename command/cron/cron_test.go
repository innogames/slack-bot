package cron

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/config"
	"github.com/innogames/slack-bot/mocks"
	"github.com/nlopes/slack"
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
		event := slack.MessageEvent{}
		event.Text = "list crons"
		slackClient.On("Reply", event, mock.MatchedBy(func(input string) bool {
			return strings.HasPrefix(input, "*1 crons:*\n - `0 0 * * *`, next in")
		}))
		actual := commands.Run(event)
		assert.Equal(t, true, actual)
	})
}
