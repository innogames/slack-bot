package jenkins

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/mocks"
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
		message := msg.Message{}
		message.Text = "retry"

		actual := command.Run(message)
		assert.Equal(t, false, actual)
	})

	t.Run("Retry not existing job", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "retry job NotExisting #3"

		slackClient.On("ReplyError", message, fmt.Errorf("job *NotExisting* is not whitelisted")).Return(true)
		actual := command.Run(message)
		assert.Equal(t, true, actual)
	})

	t.Run("Retry not existing job", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "retry job TestJob #3"

		slackClient.On("SendMessage", message, "Job *TestJob* does not exist").Return("")

		jenkins.On("GetJob", "TestJob").Return(nil, fmt.Errorf(""))
		actual := command.Run(message)
		assert.Equal(t, true, actual)
	})
}
