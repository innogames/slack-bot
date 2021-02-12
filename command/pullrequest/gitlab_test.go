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

	t.Run("get empty approvers", func(t *testing.T) {
		mr := &gitlab.MergeRequest{}
		actual := gitlabFetcher.getApprovers(mr, 1)
		assert.Equal(t, []string{}, actual)
	})
}
