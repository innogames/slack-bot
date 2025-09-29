package pullrequest

import (
	"os"
	"testing"
	"time"

	"github.com/google/go-github/github"
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/innogames/slack-bot/v2/command/queue"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGithub(t *testing.T) {
	slackClient := mocks.NewSlackClient(t)
	base := bot.BaseCommand{SlackClient: slackClient}

	cfg := &config.DefaultConfig
	cfg.Github.AccessToken = os.Getenv("BOT_GITHUB_ACCESS_TOKEN")

	commands := bot.Commands{}
	cmd := newGithubCommand(base, cfg).(command)
	githubFetcher := cmd.fetcher.(*githubFetcher)
	commands.AddCommand(cmd)

	t.Run("help", func(t *testing.T) {
		help := commands.GetHelp()
		assert.Len(t, help, 1)
	})

	t.Run("invalid PR", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "https://github.com/innogames/slack-bot/pull/40000/"

		msgRef := slack.NewRefToMessage(message.Channel, message.Timestamp)

		slackClient.On("GetReactions", msgRef, slack.NewGetReactionsParameters()).Return(nil, nil)
		mocks.AssertReaction(slackClient, "x", message)

		actual := commands.Run(message)
		queue.WaitTillHavingNoQueuedMessage()
		assert.True(t, actual)
	})

	t.Run("valid PR link", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "https://github.com/innogames/slack-bot/pull/1/"

		msgRef := slack.NewRefToMessage(message.Channel, message.Timestamp)

		slackClient.On("GetReactions", msgRef, slack.NewGetReactionsParameters()).Return(nil, nil)
		mocks.AssertReaction(slackClient, "twisted_rightwards_arrows", message)

		actual := commands.Run(message)
		time.Sleep(time.Millisecond * 200)
		assert.True(t, actual)
	})

	t.Run("get real PR", func(t *testing.T) {
		pr, err := cmd.fetcher.getPullRequest(matcher.Result{
			"project": "innogames",
			"repo":    "slack-bot",
			"number":  "1",
		}, &cfg.PullRequest)

		expected := pullRequest{
			Name:      "Add weather command",
			Author:    "pbojan",
			Link:      "https://api.github.com/repos/innogames/slack-bot/pulls/1",
			Status:    prStatusMerged,
			Approvers: []string{},
		}

		require.NoError(t, err)
		assert.Equal(t, expected, pr)
	})

	t.Run("Render template ", func(t *testing.T) {
		tpl, err := util.CompileTemplate(`{{$pr := githubPullRequest "innogames" "slack-bot" "1"}}PR: "{{$pr.Name}}"`)
		require.NoError(t, err)

		res, err := util.EvalTemplate(tpl, util.Parameters{})
		require.NoError(t, err)

		assert.Equal(t, `PR: "Add weather command"`, res)
	})

	t.Run("get status", func(t *testing.T) {
		state := "closed"

		pr := &github.PullRequest{}
		pr.State = &state
		actual := githubFetcher.getStatus(pr, true)
		assert.Equal(t, prStatusClosed, actual)

		state = "open"
		pr.State = &state
		actual = githubFetcher.getStatus(pr, true)
		assert.Equal(t, prStatusInReview, actual)

		pr.State = &state
		actual = githubFetcher.getStatus(pr, false)
		assert.Equal(t, prStatusOpen, actual)
	})
}
