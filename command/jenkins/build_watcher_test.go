package jenkins

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuildWatcher(t *testing.T) {
	slackClient := mocks.SlackClient{}
	jenkins := &mocks.Client{}

	command := bot.Commands{}
	command.AddCommand(newBuildWatcherCommand(jenkins, &slackClient))

	t.Run("Test invalid command", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "notify me not"

		actual := command.Run(message)
		assert.Equal(t, false, actual)
	})

	t.Run("build notifier with invalid job", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "notify build TestJob"

		jenkins.On("GetJob", "TestJob").Return(nil, fmt.Errorf(""))
		slackClient.On("SendMessage", message, "Job *TestJob* does not exist").Return("")
		actual := command.Run(message)
		assert.Equal(t, true, actual)
	})

	t.Run("help", func(t *testing.T) {
		help := command.GetHelp()
		assert.Equal(t, 2, len(help))
	})
}
