package jenkins

import (
	"context"
	"errors"
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testJenkinsHost = "https://jenkins.example.com"

func TestBuildWatcher(t *testing.T) {
	slackClient, jenkinsClient, base := getTestJenkinsCommand()

	command := bot.Commands{}
	command.AddCommand(newBuildWatcherCommand(base, testJenkinsHost))

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

	t.Run("build notifier with valid URL and build number", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "inform job https://jenkins.example.com/job/BuildMobileClient/36659/"

		ctx := context.Background()
		jenkinsClient.On("GetJob", ctx, "BuildMobileClient").Return(nil, errors.New(""))
		mocks.AssertSlackMessage(slackClient, message, "Job *BuildMobileClient* does not exist")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("build notifier with nested folder URL", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "inform job https://jenkins.example.com/job/Folder/job/JobName/123/"

		ctx := context.Background()
		jenkinsClient.On("GetJob", ctx, "Folder/JobName").Return(nil, errors.New(""))
		mocks.AssertSlackMessage(slackClient, message, "Job *Folder/JobName* does not exist")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("build notifier with multibranch pipeline URL", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "inform job https://jenkins.example.com/job/Pipeline/job/feature%2Fbranch/456/"

		ctx := context.Background()
		// URL-encoded %2F gets decoded to /
		jenkinsClient.On("GetJob", ctx, "Pipeline/feature/branch").Return(nil, errors.New(""))
		mocks.AssertSlackMessage(slackClient, message, "Job *Pipeline/feature/branch* does not exist")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("build notifier with URL without build number", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "inform job https://jenkins.example.com/job/BuildMobileClient/"

		ctx := context.Background()
		jenkinsClient.On("GetJob", ctx, "BuildMobileClient").Return(nil, errors.New(""))
		mocks.AssertSlackMessage(slackClient, message, "Job *BuildMobileClient* does not exist")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("build notifier with invalid host URL", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "inform job https://wrong-host.com/job/Job/123/"

		mocks.AssertSlackMessage(slackClient, message, "URL does not match configured Jenkins host")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("build notifier with deeply nested URL", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "inform job https://jenkins.example.com/job/A/job/B/job/C/789/"

		ctx := context.Background()
		jenkinsClient.On("GetJob", ctx, "A/B/C").Return(nil, errors.New(""))
		mocks.AssertSlackMessage(slackClient, message, "Job *A/B/C* does not exist")
		actual := command.Run(message)
		assert.True(t, actual)
	})
}

func TestParseJenkinsURL(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		expectedJob   string
		expectedBuild int
		expectError   bool
	}{
		{
			name:          "simple job with build",
			url:           "https://jenkins.example.com/job/BuildMobileClient/36659/",
			expectedJob:   "BuildMobileClient",
			expectedBuild: 36659,
		},
		{
			name:          "nested folder job",
			url:           "https://jenkins.example.com/job/Folder/job/JobName/123/",
			expectedJob:   "Folder/JobName",
			expectedBuild: 123,
		},
		{
			name:          "multibranch with encoded slash",
			url:           "https://jenkins.example.com/job/Pipeline/job/feature%2Fbranch/456/",
			expectedJob:   "Pipeline/feature%2Fbranch",
			expectedBuild: 456,
		},
		{
			name:          "job without build number",
			url:           "https://jenkins.example.com/job/BuildMobileClient/",
			expectedJob:   "BuildMobileClient",
			expectedBuild: 0,
		},
		{
			name:          "deeply nested",
			url:           "https://jenkins.example.com/job/A/job/B/job/C/789/",
			expectedJob:   "A/B/C",
			expectedBuild: 789,
		},
		{
			name:          "without trailing slash",
			url:           "https://jenkins.example.com/job/Job/100",
			expectedJob:   "Job",
			expectedBuild: 100,
		},
		{
			name:        "invalid URL - no job segment",
			url:         "https://jenkins.example.com/view/something/",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			job, build, err := parseJenkinsURL(tc.url)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedJob, job)
				assert.Equal(t, tc.expectedBuild, build)
			}
		})
	}
}
