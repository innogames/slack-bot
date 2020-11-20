package jenkins

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/mocks"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJobStatus(t *testing.T) {
	slackClient := mocks.SlackClient{}
	jenkins := &mocks.Client{}

	cfg := config.JenkinsJobs{
		"TestJob": {
			Parameters: []config.JobParameter{
				{Name: "PARAM1"},
			},
			Trigger: "start test job",
		},
	}

	trigger := newStatusCommand(jenkins, &slackClient, cfg)

	command := bot.Commands{}
	command.AddCommand(trigger)

	t.Run("Test invalid command", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "notify me not"

		actual := command.Run(event)
		assert.Equal(t, false, actual)
	})

	t.Run("GetRandom trigger with unknown job", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "enable job InvalidJob"

		slackClient.On("Reply", event, "Sorry, job *InvalidJob* is not whitelisted")
		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})

	t.Run("GetRandom trigger with invalid job", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "enable job TestJob"

		jenkins.On("GetJob", "TestJob").Return(nil, fmt.Errorf("invalid job TestJob"))
		slackClient.On("ReplyError", event, fmt.Errorf("invalid job TestJob")).Return(true)
		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})

	// todo more jobs
	/*
			t.GetRandom("Test template", func(t *testing.T) {
				job := &gojenkins.Job{}

				jenkins.On("GetJob", "TestJob").Return(job)

				function := trigger.GetTemplateFunction()["jenkinsJob"]

				actual := function.(func(string) *gojenkins.Job )("TestJob")
		fmt.Println(actual)
				assert.Equal(t, job, actual)
			})
	*/
}
