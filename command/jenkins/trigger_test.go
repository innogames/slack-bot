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
	"github.com/stretchr/testify/mock"
)

func TestJenkinsTrigger(t *testing.T) {
	slackClient, jenkinsClient, base := getTestJenkinsCommand()

	cfg := config.JenkinsJobs{
		"TestJob": {
			Parameters: []config.JobParameter{
				{Name: "PARAM1"},
			},
			Trigger: "start test job",
		},
		"TestJobWithTrigger": {
			Parameters: []config.JobParameter{},
			Trigger:    "just do it",
		},
		"TestJobWithoutTrigger": {
			Parameters: []config.JobParameter{},
		},
		"Prefix/Test": {
			Parameters: []config.JobParameter{
				{Name: "PARAM1"},
			},
		},
	}

	trigger := newTriggerCommand(base, cfg)

	command := bot.Commands{}
	command.AddCommand(trigger)

	t.Run("Test help", func(t *testing.T) {
		help := command.GetHelp()
		assert.Len(t, help, 3)
	})

	t.Run("Trigger not existing job", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "trigger job NotExisting"

		expectedJobs := "Prefix/Test* \n - *TestJob* \n - *TestJobWithTrigger* \n - *TestJobWithoutTrigger"
		mocks.AssertSlackMessage(slackClient, message, "Sorry, job *NotExisting* is not startable. Possible jobs: \n - *"+expectedJobs+"*")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Trigger URL-encoded job name not whitelisted", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "trigger job backend/game-backend/xandreas%2FHC-7231"

		// The decoded job name should be used in the error message
		decodedJobName := "backend/game-backend/xandreas/HC-7231"
		expectedJobs := "Prefix/Test* \n - *TestJob* \n - *TestJobWithTrigger* \n - *TestJobWithoutTrigger"
		mocks.AssertSlackMessage(slackClient, message, "Sorry, job *"+decodedJobName+"* is not startable. Possible jobs: \n - *"+expectedJobs+"*")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Trigger URL-encoded job name that is whitelisted", func(t *testing.T) {
		// Create a separate config for this test to avoid affecting other tests
		testCfg := config.JenkinsJobs{
			"TestJob": {
				Parameters: []config.JobParameter{
					{Name: "PARAM1"},
				},
				Trigger: "start test job",
			},
			"TestJobWithTrigger": {
				Parameters: []config.JobParameter{},
				Trigger:    "just do it",
			},
			"TestJobWithoutTrigger": {
				Parameters: []config.JobParameter{},
			},
			"Prefix/Test": {
				Parameters: []config.JobParameter{
					{Name: "PARAM1"},
				},
			},
			"backend/game-backend/xandreas%2FHC-7231": {
				Parameters: []config.JobParameter{
					{Name: "PARAM1"},
				},
			},
		}

		testTrigger := newTriggerCommand(base, testCfg)
		testCommand := bot.Commands{}
		testCommand.AddCommand(testTrigger)

		message := msg.Message{}
		message.Text = "trigger job backend%2Fgame-backend%2Fxandreas%252FHC-7231"

		// Expect an error about missing parameters since we're not providing PARAM1
		mocks.AssertError(slackClient, message, "sorry, you have to pass 1 parameters (PARAM1)")

		actual := testCommand.Run(message)
		assert.True(t, actual)
	})

	t.Run("Not enough parameters", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "trigger job TestJob"

		mocks.AssertError(slackClient, message, "sorry, you have to pass 1 parameters (PARAM1)")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("matched trigger, but not all parameters provided", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "start test job"

		mocks.AssertError(slackClient, message, "sorry, you have to pass 1 parameters (PARAM1)")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("matched prefixed trigger, but not all parameters provided", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "trigger job Prefix/Test"

		mocks.AssertError(slackClient, message, "sorry, you have to pass 1 parameters (PARAM1)")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("generic trigger", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "trigger job TestJob foo"

		mocks.AssertReaction(slackClient, "coffee", message)

		slackClient.On(
			"ReplyError",
			message,
			mock.Anything,
		)

		ctx := context.Background()
		jenkinsClient.On("GetJob", ctx, "TestJob").Return(nil, errors.New("404"))
		actual := command.Run(message)

		assert.True(t, actual)
	})

	t.Run("custom trigger", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "just do it"

		mocks.AssertReaction(slackClient, "coffee", message)

		slackClient.On(
			"ReplyError",
			message,
			mock.Anything,
		)

		ctx := context.Background()
		jenkinsClient.On("GetJob", ctx, "TestJobWithTrigger").Return(nil, errors.New("404"))
		actual := command.Run(message)

		assert.True(t, actual)
	})

	t.Run("No trigger found...do nothing", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "start foo job"

		actual := command.Run(message)
		assert.False(t, actual)
	})
}
