package pullrequest

import (
	"testing"
	"time"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/pkg/errors"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

type testFetcher struct {
	pr  pullRequest
	err error
}

func (t *testFetcher) getPullRequest(match matcher.Result) (pullRequest, error) {
	return t.pr, t.err
}

func (t *testFetcher) getHelp() []bot.Help {
	return []bot.Help{}
}

func TestGetCommands(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	cfg := &config.Config{}

	// as we pass a empty config, no PR fetcher is able to register -> 0 valid commands
	commands := GetCommands(base, cfg)
	assert.Equal(t, 1, commands.Count())
}

func TestPullRequest(t *testing.T) {
	slackClient := &mocks.SlackClient{}

	t.Run("invalid command", func(t *testing.T) {
		commands, _ := initTest(slackClient)

		message := msg.Message{}
		message.Text = "quatsch"

		actual := commands.Run(message)
		assert.False(t, actual)
	})

	t.Run("PR not found", func(t *testing.T) {
		commands, fetcher := initTest(slackClient)

		message := msg.Message{}
		fetcher.err = errors.New("PR not found")
		message.Text = "vcd.example.com/projects/foo/repos/bar/pull-requests/1337"

		slackClient.On("ReplyError", message, fetcher.err)
		mocks.AssertReaction(slackClient, "x", message)

		actual := commands.Run(message)
		assert.True(t, actual)
	})

	t.Run("PR got merged", func(t *testing.T) {
		commands, fetcher := initTest(slackClient)

		message := msg.Message{}
		fetcher.err = nil
		fetcher.pr = pullRequest{
			Status:    prStatusMerged,
			Approvers: []string{"test"},
		}
		message.Text = "vcd.example.com/projects/foo/repos/bar/pull-requests/1337"

		msgRef := slack.NewRefToMessage(message.Channel, message.Timestamp)
		slackClient.
			On("GetReactions", msgRef, slack.NewGetReactionsParameters()).Return(nil, nil)

		mocks.AssertRemoveReaction(slackClient, "eyes", message)
		mocks.AssertReaction(slackClient, "white_check_mark", message)
		mocks.AssertReaction(slackClient, "twisted_rightwards_arrows", message)

		actual := commands.Run(message)
		assert.True(t, actual)
		time.Sleep(time.Millisecond * 10) // todo channel
	})

	t.Run("PR got declined", func(t *testing.T) {
		commands, fetcher := initTest(slackClient)

		message := msg.Message{}
		fetcher.err = nil
		fetcher.pr = pullRequest{
			Status:    prStatusClosed,
			Approvers: []string{},
		}
		message.Text = "vcd.example.com/projects/foo/repos/bar/pull-requests/1337"

		mocks.AssertRemoveReaction(slackClient, "eyes", message)
		mocks.AssertRemoveReaction(slackClient, "white_check_mark", message)
		mocks.AssertReaction(slackClient, "x", message)

		actual := commands.Run(message)
		assert.True(t, actual)
		time.Sleep(time.Millisecond * 10) // todo channel
	})

	t.Run("PR got Approvers", func(t *testing.T) {
		commands, fetcher := initTest(slackClient)

		message := msg.Message{}
		fetcher.err = nil
		fetcher.pr = pullRequest{
			Status:    prStatusOpen,
			Approvers: []string{"test"},
		}
		message.Text = "vcd.example.com/projects/foo/repos/bar/pull-requests/1337"

		mocks.AssertRemoveReaction(slackClient, "eyed", message)
		mocks.AssertRemoveReaction(slackClient, "x", message)
		mocks.AssertReaction(slackClient, "white_check_mark", message)

		actual := commands.Run(message)
		assert.True(t, actual)
		time.Sleep(time.Millisecond * 10) // todo channel
	})

	t.Run("PR in review", func(t *testing.T) {
		commands, fetcher := initTest(slackClient)

		message := msg.Message{}
		fetcher.err = nil
		fetcher.pr = pullRequest{
			Status:    prStatusInReview,
			Approvers: []string{},
		}
		message.Text = "vcd.example.com/projects/foo/repos/bar/pull-requests/1337"

		mocks.AssertReaction(slackClient, "eyes", message)

		actual := commands.Run(message)
		assert.True(t, actual)
		time.Sleep(time.Millisecond * 10) // todo channel
	})
}

// creates a fresh instance of Commands a clean Fetcher to avoid racing conditions
func initTest(slackClient client.SlackClient) (bot.Commands, *testFetcher) {
	fetcher := &testFetcher{}
	commands := bot.Commands{}
	base := bot.BaseCommand{SlackClient: slackClient}

	cfg := config.DefaultConfig

	cmd := &command{
		base,
		cfg.PullRequest,
		fetcher,
		".*/projects/(?P<project>.+)/repos/(?P<repo>.+)/pull-requests/(?P<number>\\d+).*",
	}
	commands.AddCommand(cmd)

	return commands, fetcher
}
