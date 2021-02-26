package pullrequest

import (
	"github.com/xanzy/go-gitlab"
	"testing"

	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/mocks"
	"github.com/stretchr/testify/assert"
)

func TestGitlab(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	cfg := &config.Config{}
	cfg.Gitlab.AccessToken = "https://gitlab.example.com"
	cfg.Gitlab.AccessToken = "0815"

	commands := bot.Commands{}
	cmd := newGitlabCommand(base, cfg).(command)
	gitlabFetcher := cmd.fetcher.(*gitlabFetcher)

	commands.AddCommand(cmd)

	t.Run("help", func(t *testing.T) {
		help := commands.GetHelp()
		assert.Equal(t, 1, len(help))
	})

	t.Run("get status", func(t *testing.T) {
		mr := &gitlab.MergeRequest{}
		actual := gitlabFetcher.getStatus(mr)
		assert.Equal(t, prStatusOpen, actual)

		mr = &gitlab.MergeRequest{}
		mr.State = "merged"
		actual = gitlabFetcher.getStatus(mr)
		assert.Equal(t, prStatusMerged, actual)

		mr = &gitlab.MergeRequest{}
		mr.State = "closed"
		actual = gitlabFetcher.getStatus(mr)
		assert.Equal(t, prStatusClosed, actual)
	})

	t.Run("test convertToPullRequest", func(t *testing.T) {
		mr := &gitlab.MergeRequest{}
		mr.State = "open"
		mr.Pipeline = &gitlab.PipelineInfo{}
		mr.Pipeline.Status = "running"
		mr.Title = "my title"

		actual := gitlabFetcher.convertToPullRequest(mr, 100)

		expected := pullRequest{}
		expected.Status = prStatusOpen
		expected.BuildStatus = buildStatusRunning
		expected.Approvers = []string{}
		expected.Name = "my title"

		assert.Equal(t, expected, actual)
	})

	t.Run("get build status", func(t *testing.T) {
		mr := &gitlab.MergeRequest{}
		actual := gitlabFetcher.getPipelineStatus(mr)
		assert.Equal(t, buildStatusUnknown, actual)

		mr = &gitlab.MergeRequest{}
		mr.Pipeline = &gitlab.PipelineInfo{}
		actual = gitlabFetcher.getPipelineStatus(mr)
		assert.Equal(t, buildStatusUnknown, actual)

		mr = &gitlab.MergeRequest{}
		mr.Pipeline = &gitlab.PipelineInfo{}
		mr.Pipeline.Status = "failed"
		actual = gitlabFetcher.getPipelineStatus(mr)
		assert.Equal(t, buildStatusFailed, actual)

		mr = &gitlab.MergeRequest{}
		mr.Pipeline = &gitlab.PipelineInfo{}
		mr.Pipeline.Status = "success"
		actual = gitlabFetcher.getPipelineStatus(mr)
		assert.Equal(t, buildStatusSuccess, actual)
	})

	t.Run("get empty approvers", func(t *testing.T) {
		mr := &gitlab.MergeRequest{}
		actual := gitlabFetcher.getApprovers(mr, 1)
		assert.Equal(t, []string{}, actual)
	})
}
