package jenkins

import (
	"fmt"
	"testing"

	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/stretchr/testify/assert"
)

func TestJenkinsWatcher(t *testing.T) {
	slackClient, jenkins, base := getTestJenkinsCommand()

	command := bot.Commands{}
	command.AddCommand(newJobWatcherCommand(base))

	t.Run("Test watch invalid job", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "watch TestJob"

		jenkins.On("GetJob", "TestJob").Return(nil, fmt.Errorf("404"))
		slackClient.On("ReplyError", message, fmt.Errorf("404"))

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Test unwatch", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "unwatch TestJob"

		slackClient.On("SendMessage", message, "Okay, you just unwatched TestJob").Return("")

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Test help", func(t *testing.T) {
		help := command.GetHelp()
		assert.NotNil(t, help)
	})
}
