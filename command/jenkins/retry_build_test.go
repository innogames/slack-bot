package jenkins

import (
	"context"
	"fmt"
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
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

		slackClient.On("ReplyError", message, fmt.Errorf("job *NotExisting* is not whitelisted")).Return(true)
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Retry not existing job", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "retry job TestJob #3"

		slackClient.On("SendMessage", message, "Job *TestJob* does not exist").Return("")

		ctx := context.TODO()
		jenkinsClient.On("GetJob", ctx, "TestJob").Return(nil, fmt.Errorf(""))
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Test help", func(t *testing.T) {
		help := command.GetHelp()
		assert.Equal(t, 1, len(help))
	})
}
