package jenkins

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/mocks"
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuildWatcher(t *testing.T) {
	slackClient := mocks.SlackClient{}
	jenkins := &mocks.Client{}

	command := bot.Commands{}
	command.AddCommand(newBuildWatcherCommand(jenkins, &slackClient))

	t.Run("Test invalid command", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "notify me not"

		actual := command.Run(event)
		assert.Equal(t, false, actual)
	})

	t.Run("build notifier with invalid job", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "notify build TestJob"

		jenkins.On("GetJob", "TestJob").Return(nil, fmt.Errorf(""))
		slackClient.On("Reply", event, "Job *TestJob* does not exist")
		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})
}
