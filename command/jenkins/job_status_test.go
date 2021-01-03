package jenkins

import (
	"fmt"
	"testing"

	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/msg"
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

		slackClient.On("SendMessage", message, "Sorry, job *InvalidJob* is not whitelisted").Return("")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("GetRandom trigger with invalid job", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "enable job TestJob"

		jenkins.On("GetJob", "TestJob").Return(nil, fmt.Errorf("invalid job TestJob"))
		slackClient.On("ReplyError", message, fmt.Errorf("invalid job TestJob")).Return(true)
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Test help", func(t *testing.T) {
		help := command.GetHelp()
		assert.NotNil(t, help)
	})
}
