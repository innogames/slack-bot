package jenkins

import (
	"context"
	"errors"
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
)

func TestJenkinsRetry(t *testing.T) {
	slackClient, jenkinsClient, base := getTestJenkinsCommand()

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
	command.AddCommand(newRetryCommand(base, cfg))

	t.Run("Test invalid command", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "retry"

		actual := command.Run(message)
		assert.False(t, actual)
	})

	t.Run("Retry not existing job", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "retry job NotExisting #3"

		mocks.AssertError(slackClient, message, "job *NotExisting* is not whitelisted")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Retry URL-encoded job name not whitelisted", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "retry job backend/game-backend/xandreas%2FHC-7231"

		// The decoded job name should be used in the error message
		decodedJobName := "backend/game-backend/xandreas/HC-7231"
		mocks.AssertError(slackClient, message, "job *"+decodedJobName+"* is not whitelisted")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Retry URL-encoded job name that is whitelisted", func(t *testing.T) {
		// Add a URL-encoded job to the config for testing
		cfg["backend/game-backend/xandreas%2FHC-7231"] = config.JobConfig{
			Parameters: []config.JobParameter{
				{Name: "PARAM1"},
			},
		}

		message := msg.Message{}
		message.Text = "retry job backend%2Fgame-backend%2Fxandreas%252FHC-7231"

		ctx := context.TODO()
		// The decoded job name should be passed to Jenkins
		decodedJobName := "backend/game-backend/xandreas%2FHC-7231"
		jenkinsClient.On("GetJob", ctx, decodedJobName).Return(nil, errors.New("job not found"))
		mocks.AssertSlackMessage(slackClient, message, "Job *"+decodedJobName+"* does not exist")

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Retry not existing job", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "retry job TestJob #3"

		mocks.AssertSlackMessage(slackClient, message, "Job *TestJob* does not exist")

		ctx := context.TODO()
		jenkinsClient.On("GetJob", ctx, "TestJob").Return(nil, errors.New(""))
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Test help", func(t *testing.T) {
		help := command.GetHelp()
		assert.Len(t, help, 1)
	})
}
