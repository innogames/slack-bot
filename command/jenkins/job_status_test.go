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

func TestJobStatus(t *testing.T) {
	slackClient, jenkins, base := getTestJenkinsCommand()

	cfg := config.JenkinsJobs{
		"TestJob": {
			Parameters: []config.JobParameter{
				{Name: "PARAM1"},
			},
			Trigger: "start test job",
		},
	}

	trigger := newStatusCommand(base, cfg)

	command := bot.Commands{}
	command.AddCommand(trigger)

	t.Run("Test invalid command", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "notify me not"

		actual := command.Run(message)
		assert.False(t, actual)
	})

	t.Run("GetRandom trigger with unknown job", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "enable job InvalidJob"

		mocks.AssertSlackMessage(slackClient, message, "Sorry, job *InvalidJob* is not whitelisted")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Enable URL-encoded job name not whitelisted", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "enable job backend/game-backend/xandreas%2FHC-7231"

		// The decoded job name should be used in the error message
		decodedJobName := "backend/game-backend/xandreas/HC-7231"
		mocks.AssertSlackMessage(slackClient, message, "Sorry, job *"+decodedJobName+"* is not whitelisted")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Disable URL-encoded job name that is whitelisted", func(t *testing.T) {
		// Add a URL-encoded job to the config for testing
		cfg["backend/game-backend/xandreas%2FHC-7231"] = config.JobConfig{
			Parameters: []config.JobParameter{
				{Name: "PARAM1"},
			},
		}

		message := msg.Message{}
		message.Text = "disable job backend%2Fgame-backend%2Fxandreas%252FHC-7231"

		ctx := context.TODO()
		// The decoded job name should be passed to Jenkins
		decodedJobName := "backend/game-backend/xandreas%2FHC-7231"
		jenkins.On("GetJob", ctx, decodedJobName).Return(nil, errors.New("job not found"))
		mocks.AssertError(slackClient, message, "job not found")

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("GetRandom trigger with invalid job", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "enable job TestJob"

		ctx := context.TODO()
		jenkins.On("GetJob", ctx, "TestJob").Return(nil, errors.New("invalid job TestJob"))
		mocks.AssertError(slackClient, message, "invalid job TestJob")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Test help", func(t *testing.T) {
		help := command.GetHelp()
		assert.NotNil(t, help)
	})
}
