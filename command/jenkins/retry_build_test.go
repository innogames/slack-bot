package jenkins

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/config"
	"github.com/innogames/slack-bot/mocks"
	"github.com/nlopes/slack"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJenkinsRetry(t *testing.T) {
	slackClient := mocks.SlackClient{}
	jenkins := &mocks.Client{}
	logger := logrus.New()
	cfg := config.JenkinsJobs{
		"TestJob": {
			Parameters: []config.JobParameter{
				{Name: "PARAM1"},
			},
			Trigger: "start test job",
		},
		"TestJobWithoutTrigger": {
			Parameters: []config.JobParameter{
				{Name: "PARAM1"},
			},
		},
	}

	command := bot.Commands{}
	command.AddCommand(newRetryCommand(jenkins, &slackClient, cfg, logger))

	t.Run("Test invalid command", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "retry"

		actual := command.Run(event)
		assert.Equal(t, false, actual)
	})

	t.Run("Retry not existing job", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "retry job NotExisting #3"

		slackClient.On("ReplyError", event, fmt.Errorf("job *NotExisting* is not whitelisted")).Return(true)
		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})

	t.Run("Retry not existing job", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "retry job TestJob #3"

		slackClient.On("Reply", event, "Job *TestJob* does not exist")

		jenkins.On("GetJob", "TestJob").Return(nil, fmt.Errorf(""))
		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})
}
