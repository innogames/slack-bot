package jenkins

import (
	"context"
	"errors"
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/mocks"
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

		ctx := context.Background()
		jenkinsClient.On("GetJob", ctx, "TestJob").Return(nil, errors.New(""))
		mocks.AssertSlackMessage(slackClient, message, "Job *TestJob* does not exist")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("build notifier with URL-encoded job name", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "notify build backend/game-backend/xandreas%2FHC-7231"

		ctx := context.Background()
		// The decoded job name should be passed to Jenkins
		decodedJobName := "backend/game-backend/xandreas/HC-7231"
		jenkinsClient.On("GetJob", ctx, decodedJobName).Return(nil, errors.New(""))
		mocks.AssertSlackMessage(slackClient, message, "Job *backend/game-backend/xandreas/HC-7231* does not exist")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("build notifier with double URL-encoded job name", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "notify build backend/game-backend/xandreas%252FHC-7231"

		ctx := context.Background()
		// The decoded job name should be passed to Jenkins (%252F becomes %2F, then /)
		decodedJobName := "backend/game-backend/xandreas%2FHC-7231"
		jenkinsClient.On("GetJob", ctx, decodedJobName).Return(nil, errors.New(""))
		mocks.AssertSlackMessage(slackClient, message, "Job *backend/game-backend/xandreas%2FHC-7231* does not exist")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("build notifier with special characters in job name", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "notify build backend/game-backend/feature%2Fbranch%2Btest"

		ctx := context.Background()
		// The decoded job name should handle multiple URL-encoded characters
		decodedJobName := "backend/game-backend/feature/branch+test"
		jenkinsClient.On("GetJob", ctx, decodedJobName).Return(nil, errors.New(""))
		mocks.AssertSlackMessage(slackClient, message, "Job *backend/game-backend/feature/branch+test* does not exist")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("help", func(t *testing.T) {
		help := command.GetHelp()
		assert.Len(t, help, 2)
	})
}
