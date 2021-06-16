package jenkins

import (
	"context"
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJenkinsWatcher(t *testing.T) {
	slackClient, jenkinsClient, base := getTestJenkinsCommand()

	command := bot.Commands{}
	command.AddCommand(newJobWatcherCommand(base))

	t.Run("Test watch invalid job", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "watch TestJob"

		ctx := context.TODO()
		jenkinsClient.On("GetJob", ctx, "TestJob").Return(nil, fmt.Errorf("404"))
		mocks.AssertError(slackClient, message, "404")

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Test unwatch", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "unwatch TestJob"

		mocks.AssertSlackMessage(slackClient, message, "Okay, you just unwatched TestJob")

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Test help", func(t *testing.T) {
		help := command.GetHelp()
		assert.NotNil(t, help)
	})
}
