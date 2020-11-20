package jenkins

import (
	"errors"
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/mocks"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestJenkinsTrigger(t *testing.T) {
	slackClient := mocks.SlackClient{}
	jenkinsClient := &mocks.Client{}
	logger := logrus.New()
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

	trigger := newTriggerCommand(jenkinsClient, &slackClient, cfg, logger)

	command := bot.Commands{}
	command.AddCommand(trigger)

	t.Run("Test help", func(t *testing.T) {
		help := command.GetHelp()
		assert.Equal(t, 3, len(help))
	})

	t.Run("Trigger not existing job", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "trigger job NotExisting"

		slackClient.On("Reply", event, "Sorry, job *NotExisting* is not startable. Possible jobs: \n - *Prefix/Test* \n - *TestJob* \n - *TestJobWithTrigger* \n - *TestJobWithoutTrigger*")
		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})

	t.Run("Not enough parameters", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "trigger job TestJob"

		slackClient.On("ReplyError", event, fmt.Errorf("sorry, you have to pass 1 parameters (PARAM1)"))
		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})

	t.Run("matched trigger, but not all parameters provided", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "start test job"

		slackClient.On("ReplyError", event, fmt.Errorf("sorry, you have to pass 1 parameters (PARAM1)"))
		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})

	t.Run("matched prefixed trigger, but not all parameters provided", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "trigger job Prefix/Test"

		slackClient.On("ReplyError", event, fmt.Errorf("sorry, you have to pass 1 parameters (PARAM1)"))
		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})

	t.Run("generic trigger", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "trigger job TestJob foo"

		slackClient.On(
			"AddReaction",
			"coffee",
			slack.NewRefToMessage(event.Channel, event.Timestamp),
		)

		slackClient.On(
			"ReplyError",
			event,
			mock.Anything,
		)

		jenkinsClient.On("GetJob", "TestJob").Return(nil, errors.New("404"))
		actual := command.Run(event)

		assert.Equal(t, true, actual)
	})

	t.Run("custom trigger", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "just do it"

		slackClient.On(
			"AddReaction",
			"coffee",
			slack.NewRefToMessage(event.Channel, event.Timestamp),
		)

		slackClient.On(
			"ReplyError",
			event,
			mock.Anything,
		)

		jenkinsClient.On("GetJob", "TestJobWithTrigger").Return(nil, errors.New("404"))
		actual := command.Run(event)

		assert.Equal(t, true, actual)
	})

	t.Run("No trigger found...do nothing", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "start foo job"

		actual := command.Run(event)
		assert.Equal(t, false, actual)
	})
}
