package pullrequest

import (
	"testing"
	"time"

	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/mocks"
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
		slackClient.On("AddReaction", "x", message)

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

		slackClient.On("RemoveReaction", "eyes", message)
		slackClient.On("AddReaction", "white_check_mark", message)
		slackClient.On("AddReaction", "twisted_rightwards_arrows", message)

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

		slackClient.On("RemoveReaction", "eyes", message)
		slackClient.On("RemoveReaction", "white_check_mark", message)
		slackClient.On("AddReaction", "x", message)

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

		slackClient.On("RemoveReaction", "eyes", message)
		slackClient.On("RemoveReaction", "x", message)
		slackClient.On("AddReaction", "white_check_mark", message)

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

		slackClient.On("AddReaction", "eyes", message)

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
