package jira

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCommentJira(t *testing.T) {
	slackClient := &mocks.SlackClient{}

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	// mock TEST-1234 as ticket
	mux.HandleFunc("/rest/api/2/issue/TEST-1234", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte(`{
			"key": "TEST-1234"
		}`))
	})

	cfg := &config.Jira{
		Host:    server.URL,
		Project: "TEST",
	}
	jiraClient, err := client.GetJiraClient(cfg)
	assert.Nil(t, err)

	command := bot.Commands{}
	command.AddCommand(newCommentCommand(jiraClient, slackClient, cfg))

	t.Run("No match", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "quatsch"

		actual := command.Run(message)
		assert.False(t, actual)
	})

	t.Run("comment on not existing ticket", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "add comment to ticket TEST-1235 that's true!"

		mocks.AssertError(slackClient, message, "invalid ticket: TEST-1235")

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("comment on existing ticket", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "add comment to ticket TEST-1234 that's true!"

		mocks.AssertReaction(slackClient, "✅", message)

		mux.HandleFunc("/rest/api/2/issue/TEST-1234/comment", func(res http.ResponseWriter, req *http.Request) {
			assert.Equal(t, "POST", req.Method)
			res.WriteHeader(201)
			res.Write([]byte(`{}`))
		})

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Test help", func(t *testing.T) {
		help := command.GetHelp()
		assert.Equal(t, 1, len(help))
	})
}
