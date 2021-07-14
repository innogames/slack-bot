package pullrequest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/innogames/slack-bot/v2/command/queue"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestBitbucketNotActive(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	cfg := &config.DefaultConfig

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

	server := spawnBitbucketTestServer()
	defer server.Close()

	cfg := config.DefaultConfig
	cfg.Bitbucket = config.Bitbucket{
		Host:       server.URL,
		Project:    "myProject",
		Repository: "myRepo",
		APIKey:     "0815",
	}

	command := bot.Commands{}
	cmd := newBitbucketCommand(base, &cfg)
	command.AddCommand(cmd)

	t.Run("Merged PR", func(t *testing.T) {
		message := msg.Message{}
		message.Text = server.URL + "/projects/myProject/repos/myRepo/pull-requests/1339 please review ASAP!"

		slackClient.On("GetReactions", message.GetMessageRef(), slack.NewGetReactionsParameters()).Return([]slack.ItemReaction{}, nil)
		mocks.AssertReaction(slackClient, "white_check_mark", message)
		mocks.AssertReaction(slackClient, "twisted_rightwards_arrows", message)

		actual := command.Run(message)
		assert.True(t, actual)
		time.Sleep(time.Millisecond * 200)
		assert.Equal(t, 1, queue.CountCurrentJobs())
	})

	t.Run("Test help when bitbucket is disabled", func(t *testing.T) {
		help := command.GetHelp()
		assert.Equal(t, 1, len(help))
	})

	t.Run("Render template", func(t *testing.T) {
		tpl, err := util.CompileTemplate(`{{$pr := bitbucketPullRequest "myProject" "myRepo" "1337"}}PR: {{$pr.Name}} - {{$pr.BuildStatus}}`)
		assert.Nil(t, err)

		res, err := util.EvalTemplate(tpl, util.Parameters{})
		assert.Nil(t, err)

		assert.Equal(t, fmt.Sprintf("PR: test - %d", buildStatusFailed), res)
	})

	t.Run("Render template with not existing PR", func(t *testing.T) {
		tpl, err := util.CompileTemplate(`{{$pr := bitbucketPullRequest "myProject" "myRepo" "1338"}}{{$pr.Status}} - {{$pr.BuildStatus}}`)
		assert.Nil(t, err)

		res, err := util.EvalTemplate(tpl, util.Parameters{})
		assert.Nil(t, err)

		assert.Equal(t, fmt.Sprintf("%d - %d", closedPr.Status, buildStatusUnknown), res)
	})

	t.Run("Render template with not open PR", func(t *testing.T) {
		tpl, err := util.CompileTemplate(`{{$pr := bitbucketPullRequest "myProject" "myRepo" "1339"}}{{$pr.Name}}: {{$pr.Status}} - {{$pr.BuildStatus}}`)
		assert.Nil(t, err)

		res, err := util.EvalTemplate(tpl, util.Parameters{})
		assert.Nil(t, err)

		assert.Equal(t, fmt.Sprintf("test: %d - %d", prStatusOpen, buildStatusSuccess), res)
	})
}

func spawnBitbucketTestServer() *httptest.Server {
	mux := http.NewServeMux()

	// 1337: merged pr
	mux.HandleFunc("/rest/api/1.0/projects/myProject/repos/myRepo/pull-requests/1337", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte(`{
			"id": 1337,
			"title": "test",
			"state": "MERGED",
			"reviewers": [{
				"user": {
					"name": "John Doe"
				},
				"approved": true
			}],
			"fromRef": {
				"latestCommit": "commitWithFailedBuild"
			}
		}`))
	})

	// 1339: open PR
	mux.HandleFunc("/rest/api/1.0/projects/myProject/repos/myRepo/pull-requests/1339", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte(`{
			"id": 1339,
			"title": "test",
			"state": "CLOSED",
			"reviewers": [],
			"fromRef": {
				"latestCommit": "commitWithSuccessfulBuild"
			}
		}`))
	})

	// successful build
	mux.HandleFunc("/rest/build-status/1.0/commits/commitWithSuccessfulBuild", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte(`{
			"values": [{
				"state": "SUCCESS"
			}]
		}`))
	})

	// failed build
	mux.HandleFunc("/rest/build-status/1.0/commits/commitWithFailedBuild", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte(`{
			"values": [{
				"state": "FAILED"
			}]
		}`))
	})

	return httptest.NewServer(mux)
}
