package jenkins

import (
	"fmt"
	"testing"

	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/stretchr/testify/assert"
)

func TestBuildWatcher(t *testing.T) {
	slackClient, jenkinsClient, base := getTestJenkinsCommand()

	command := bot.Commands{}
	command.AddCommand(newBuildWatcherCommand(base))

	t.Run("Test invalid command", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "notify me not"

		actual := command.Run(message)
		assert.False(t, actual)
	})

	t.Run("build notifier with invalid job", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "notify build TestJob"

		jenkinsClient.On("GetJob", "TestJob").Return(nil, fmt.Errorf(""))
		slackClient.On("SendMessage", message, "Job *TestJob* does not exist").Return("")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("help", func(t *testing.T) {
		help := command.GetHelp()
		assert.Equal(t, 2, len(help))
	})
}
