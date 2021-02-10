package pullrequest

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/innogames/slack-bot/mocks"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestGithub(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	cfg := &config.DefaultConfig
	cfg.Github.AccessToken = os.Getenv("BOT_GITHUB_ACCESS_TOKEN")

	commands := bot.Commands{}
	cmd := newGithubCommand(base, cfg).(command)
	commands.AddCommand(cmd)

	t.Run("help", func(t *testing.T) {
		help := commands.GetHelp()
		assert.Equal(t, 1, len(help))
	})

	t.Run("invalid PR", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "https://github.com/innogames/slack-bot/pull/40000/"

		msgRef := slack.NewRefToMessage(message.Channel, message.Timestamp)

		slackClient.On("GetReactions", msgRef, slack.NewGetReactionsParameters()).Return(nil, nil)
		mocks.AssertReaction(slackClient, "x", message)

		actual := commands.Run(message)
		time.Sleep(time.Millisecond * 300)
		assert.Equal(t, true, actual)
	})

	t.Run("valid PR link", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "https://github.com/innogames/slack-bot/pull/1/"

		msgRef := slack.NewRefToMessage(message.Channel, message.Timestamp)

		slackClient.On("GetReactions", msgRef, slack.NewGetReactionsParameters()).Return(nil, nil)
		mocks.AssertReaction(slackClient, "twisted_rightwards_arrows", message)

		actual := commands.Run(message)
		time.Sleep(time.Millisecond * 300)
		assert.Equal(t, true, actual)
	})

	t.Run("get real PR", func(t *testing.T) {
		pr, err := cmd.fetcher.getPullRequest(matcher.MapResult{
			"project": "innogames",
			"repo":    "slack-bot",
			"number":  "1",
		})

		expected := pullRequest{
			Name:      "Add weather command",
			Status:    prStatusMerged,
			Approvers: []string{},
		}

		assert.Nil(t, err)
		assert.Equal(t, expected, pr)
	})

	t.Run("Render template ", func(t *testing.T) {
		tpl, err := util.CompileTemplate(`{{$pr := githubPullRequest "innogames" "slack-bot" "1"}}PR: "{{$pr.Name}}"`)
		assert.Nil(t, err)

		res, err := util.EvalTemplate(tpl, util.Parameters{})
		assert.Nil(t, err)

		assert.Equal(t, `PR: "Add weather command"`, res)
	})
}
