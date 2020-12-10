package pullrequest

import (
	bitbucketServer "github.com/gfleury/go-bitbucket-v1/test/bb-mock-server/go"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/innogames/slack-bot/command/queue"
	"github.com/innogames/slack-bot/mocks"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBitbucketNotActive(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	cfg := config.Config{}

	command := bot.Commands{}
	cmd := newBitbucketCommand(base, cfg)
	command.AddCommand(cmd)

	t.Run("invalid command", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "quatsch"

		actual := command.Run(message)
		assert.False(t, actual)
	})

	t.Run("Test help when bitbucket is disabled", func(t *testing.T) {
		help := command.GetHelp()
		assert.Equal(t, 0, len(help))
	})
}

func TestBitbucketFakeServer(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	go bitbucketServer.RunServer(7992)

	// todo(matze) defer shutdown
	// todo(matze) wait till server active

	time.Sleep(time.Millisecond * 200)

	cfg := config.Bitbucket{
		Host:       "http://localhost:7992",
		Project:    "myProject",
		Repository: "myRepo",
		APIKey:     "0815",
	}

	command := bot.Commands{}
	cmd := newBitbucketCommand(base, config.Config{
		Bitbucket: cfg,
	})
	command.AddCommand(cmd)

	t.Run("Merged PR", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "http://localhost:7992/projects/myProject/repos/myRepo/pull-requests/1337 please review ASAP!"

		slackClient.On("GetReactions", message.GetMessageRef(), slack.NewGetReactionsParameters()).Return([]slack.ItemReaction{}, nil)
		slackClient.On("AddReaction", "white_check_mark", message).Once()
		slackClient.On("AddReaction", "twisted_rightwards_arrows", message).Once()

		actual := command.Run(message)
		assert.True(t, actual)
		time.Sleep(time.Millisecond * 200)
		assert.Equal(t, 0, queue.CountCurrentJobs())
	})

	t.Run("Test help when bitbucket is disabled", func(t *testing.T) {
		help := command.GetHelp()
		assert.Equal(t, 1, len(help))
	})

	t.Run("Render template ", func(t *testing.T) {
		tpl, err := util.CompileTemplate(`{{$pr := bitbucketPullRequest "myProject" "myRepo" "1337"}}PR: {{$pr.Name}}`)
		assert.Nil(t, err)

		res, err := util.EvalTemplate(tpl, util.Parameters{})
		assert.Nil(t, err)

		assert.Equal(t, "PR: test", res)
	})
}
