package jenkins

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/innogames/slack-bot/v2/mocks"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
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
		assert.Equal(t, 3, len(help))
	})

	t.Run("Trigger not existing job", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "trigger job NotExisting"

		mocks.AssertSlackMessage(slackClient, message, "Sorry, job *NotExisting* is not startable. Possible jobs: \n - *Prefix/Test* \n - *TestJob* \n - *TestJobWithTrigger* \n - *TestJobWithoutTrigger*")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Not enough parameters", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "trigger job TestJob"

		slackClient.On("ReplyError", message, fmt.Errorf("sorry, you have to pass 1 parameters (PARAM1)")).Return("")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("matched trigger, but not all parameters provided", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "start test job"

		slackClient.On("ReplyError", message, fmt.Errorf("sorry, you have to pass 1 parameters (PARAM1)")).Return("")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("matched prefixed trigger, but not all parameters provided", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "trigger job Prefix/Test"

		slackClient.On("ReplyError", message, fmt.Errorf("sorry, you have to pass 1 parameters (PARAM1)")).Return("")
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

		ctx := context.TODO()
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

		ctx := context.TODO()
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
