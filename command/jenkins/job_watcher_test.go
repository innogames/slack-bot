package jenkins

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/mocks"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJenkinsWatcher(t *testing.T) {
	slackClient := mocks.SlackClient{}
	jenkins := &mocks.Client{}
	logger := logrus.New()

	command := bot.Commands{}
	command.AddCommand(newJobWatcherCommand(jenkins, &slackClient, logger))

	t.Run("Test watch invalid job", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "watch TestJob"

		jenkins.On("GetJob", "TestJob").Return(nil, fmt.Errorf("404"))
		slackClient.On("ReplyError", event, fmt.Errorf("404"))

		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})

	t.Run("Test unwatch", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "unwatch TestJob"

		slackClient.On("Reply", event, "Okay, you just unwatched TestJob")

		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})

	t.Run("Test help", func(t *testing.T) {
		help := command.GetHelp()
		assert.NotNil(t, help)
	})
}
