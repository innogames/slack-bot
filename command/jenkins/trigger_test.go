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
		"TestJobWithoutTrigger": {
			Parameters: []config.JobParameter{
				{Name: "PARAM1"},
			},
		},
	}

	trigger := NewTriggerCommand(jenkinsClient, &slackClient, cfg, logger)

	command := bot.Commands{}
	command.AddCommand(trigger)

	t.Run("Test help", func(t *testing.T) {
		help := command.GetHelp()
		assert.Equal(t, len(cfg), len(help))
	})

	t.Run("Trigger not existing job", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "trigger job NotExisting"

		slackClient.On("Reply", event, "Sorry, job *NotExisting* is not startable. Possible jobs: \n - *TestJob* \n - *TestJobWithoutTrigger*")
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

	t.Run("GetRandom trigger, but not all parameters provided", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "start test job"

		slackClient.On("ReplyError", event, fmt.Errorf("sorry, you have to pass 1 parameters (PARAM1)"))
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
