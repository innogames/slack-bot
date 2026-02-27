package gitlab

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

const testGitlabHost = "https://gitlab.example.com"

type mockGitlabAPI struct {
	mock.Mock
}

func (m *mockGitlabAPI) GetPipeline(pid any, pipeline int64) (*gitlab.Pipeline, error) {
	args := m.Called(pid, pipeline)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gitlab.Pipeline), args.Error(1)
}

func (m *mockGitlabAPI) ListPipelineJobs(pid any, pipeline int64) ([]*gitlab.Job, error) {
	args := m.Called(pid, pipeline)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*gitlab.Job), args.Error(1)
}

func (m *mockGitlabAPI) GetJob(pid any, jobID int64) (*gitlab.Job, error) {
	args := m.Called(pid, jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gitlab.Job), args.Error(1)
}

func getTestCommand() (*mocks.SlackClient, *mockGitlabAPI, bot.Commands) {
	slackClient := &mocks.SlackClient{}
	api := &mockGitlabAPI{}

	base := gitlabCommand{
		BaseCommand: bot.BaseCommand{SlackClient: slackClient},
		api:         api,
		host:        testGitlabHost,
	}

	commands := bot.Commands{}
	commands.AddCommand(newNotifyCommand(base))

	return slackClient, api, commands
}

func TestParseGitlabURL(t *testing.T) {
	tests := []struct {
		name            string
		url             string
		expectedProject string
		expectedKind    urlType
		expectedID      int
		expectError     bool
	}{
		{
			name:            "pipeline URL",
			url:             "https://gitlab.example.com/sysadmins/game_scripts/-/pipelines/134007",
			expectedProject: "sysadmins/game_scripts",
			expectedKind:    urlTypePipeline,
			expectedID:      134007,
		},
		{
			name:            "job URL",
			url:             "https://gitlab.example.com/sysadmins/game_scripts/-/jobs/1114694",
			expectedProject: "sysadmins/game_scripts",
			expectedKind:    urlTypeJob,
			expectedID:      1114694,
		},
		{
			name:            "pipeline URL with trailing slash",
			url:             "https://gitlab.example.com/group/project/-/pipelines/100/",
			expectedProject: "group/project",
			expectedKind:    urlTypePipeline,
			expectedID:      100,
		},
		{
			name:            "nested group project",
			url:             "https://gitlab.example.com/group/subgroup/project/-/pipelines/42",
			expectedProject: "group/subgroup/project",
			expectedKind:    urlTypePipeline,
			expectedID:      42,
		},
		{
			name:            "job URL with nested groups",
			url:             "https://gitlab.example.com/a/b/c/project/-/jobs/999",
			expectedProject: "a/b/c/project",
			expectedKind:    urlTypeJob,
			expectedID:      999,
		},
		{
			name:        "invalid URL - no pipelines or jobs",
			url:         "https://gitlab.example.com/group/project/-/merge_requests/1",
			expectError: true,
		},
		{
			name:        "invalid URL - non-numeric ID",
			url:         "https://gitlab.example.com/group/project/-/pipelines/abc",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := parseGitlabURL(tc.url)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedProject, parsed.project)
				assert.Equal(t, tc.expectedKind, parsed.kind)
				assert.Equal(t, tc.expectedID, parsed.id)
			}
		})
	}
}

func TestNotifyCommand(t *testing.T) {
	t.Run("invalid command does not match", func(t *testing.T) {
		_, _, commands := getTestCommand()

		message := msg.Message{}
		message.Text = "notify pipeline something"

		actual := commands.Run(message)
		assert.False(t, actual)
	})

	t.Run("wrong host", func(t *testing.T) {
		slackClient, _, commands := getTestCommand()

		message := msg.Message{}
		message.Text = "gitlab notify https://wrong-host.com/group/project/-/pipelines/123"

		mocks.AssertSlackMessage(slackClient, message, "URL does not match configured GitLab host")
		actual := commands.Run(message)
		assert.True(t, actual)
	})

	t.Run("invalid URL path", func(t *testing.T) {
		slackClient, _, commands := getTestCommand()

		message := msg.Message{}
		message.Text = "gitlab notify https://gitlab.example.com/group/project/-/merge_requests/1"

		mocks.AssertSlackMessage(slackClient, message, "Invalid GitLab URL: URL must contain /-/pipelines/<id> or /-/jobs/<id>")
		actual := commands.Run(message)
		assert.True(t, actual)
	})

	t.Run("pipeline already finished", func(t *testing.T) {
		slackClient, api, commands := getTestCommand()

		message := msg.Message{}
		message.Text = "gitlab notify https://gitlab.example.com/group/project/-/pipelines/100"

		api.On("GetPipeline", "group/project", int64(100)).Return(&gitlab.Pipeline{
			Status: "success",
		}, nil)
		api.On("ListPipelineJobs", "group/project", int64(100)).Return([]*gitlab.Job{}, nil)

		mocks.AssertSlackMessage(slackClient, message, "GitLab pipeline *group/project/100* already finished with status: *success*")
		actual := commands.Run(message)
		assert.True(t, actual)
	})

	t.Run("job already finished", func(t *testing.T) {
		slackClient, api, commands := getTestCommand()

		message := msg.Message{}
		message.Text = "gitlab notify https://gitlab.example.com/group/project/-/jobs/200"

		api.On("GetJob", "group/project", int64(200)).Return(&gitlab.Job{
			Status: "failed",
			Name:   "test",
			Stage:  "test",
		}, nil)

		mocks.AssertSlackMessage(slackClient, message, "GitLab job *group/project/200* already finished with status: *failed*")
		actual := commands.Run(message)
		assert.True(t, actual)
	})

	t.Run("help", func(t *testing.T) {
		_, _, commands := getTestCommand()

		help := commands.GetHelp()
		assert.NotEmpty(t, help)
	})
}

func TestBuildJobSummary(t *testing.T) {
	jobs := []*gitlab.Job{
		{Status: "running"},
		{Status: "running"},
		{Status: "success"},
		{Status: "success"},
		{Status: "success"},
		{Status: "failed"},
		{Status: "pending"},
	}

	summary := buildJobSummary(jobs)
	assert.Equal(t, "2 running, 3 success, 1 failed, 1 pending", summary)
}

func TestBuildJobSummaryEmpty(t *testing.T) {
	summary := buildJobSummary(nil)
	assert.Empty(t, summary)
}

func TestIsTerminalStatus(t *testing.T) {
	assert.True(t, isTerminalStatus("success"))
	assert.True(t, isTerminalStatus("failed"))
	assert.True(t, isTerminalStatus("canceled"))
	assert.True(t, isTerminalStatus("skipped"))
	assert.True(t, isTerminalStatus("manual"))
	assert.False(t, isTerminalStatus("running"))
	assert.False(t, isTerminalStatus("pending"))
	assert.False(t, isTerminalStatus("created"))
}

func TestCountDoneJobs(t *testing.T) {
	jobs := []*gitlab.Job{
		{Status: "success"},
		{Status: "failed"},
		{Status: "running"},
		{Status: "pending"},
		{Status: "canceled"},
	}

	assert.Equal(t, 3, countDoneJobs(jobs))
	assert.Equal(t, 0, countDoneJobs(nil))
}

func TestCollectNotableJobs(t *testing.T) {
	jobs := []*gitlab.Job{
		{Name: "build", Stage: "build", Status: "success", WebURL: "https://gitlab.example.com/jobs/1"},
		{Name: "test", Stage: "test", Status: "running", WebURL: "https://gitlab.example.com/jobs/2"},
		{Name: "lint", Stage: "test", Status: "failed", WebURL: "https://gitlab.example.com/jobs/3"},
		{Name: "deploy", Stage: "deploy", Status: "pending", WebURL: "https://gitlab.example.com/jobs/4"},
	}

	details := collectNotableJobs(jobs)
	assert.Len(t, details, 2)
	assert.Equal(t, "test", details[0].name)
	assert.Equal(t, "running", details[0].status)
	assert.Equal(t, "lint", details[1].name)
	assert.Equal(t, "failed", details[1].status)
}

func TestFormatJobDetails(t *testing.T) {
	details := []jobDetail{
		{name: "test", stage: "test", status: "running", webURL: "https://gitlab.example.com/jobs/2"},
		{name: "lint", stage: "test", status: "failed", webURL: "https://gitlab.example.com/jobs/3"},
	}

	result := formatJobDetails(details)
	assert.Contains(t, result, ":arrow_forward: <https://gitlab.example.com/jobs/2|test> (test)")
	assert.Contains(t, result, ":x: <https://gitlab.example.com/jobs/3|lint> (test)")
}

func TestFormatJobDetailsEmpty(t *testing.T) {
	assert.Empty(t, formatJobDetails(nil))
}

func TestURLTypeString(t *testing.T) {
	assert.Equal(t, "pipeline", urlTypePipeline.String())
	assert.Equal(t, "job", urlTypeJob.String())
}
