package jira

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/config"
	"github.com/innogames/slack-bot/mocks"
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestCommentJira(t *testing.T) {
	slackClient := &mocks.SlackClient{}

	// todo fake http client
	cfg := config.Jira{
		Host:    "https://issues.apache.org/jira/",
		Project: "ZOOKEEPER",
	}
	jiraClient, err := client.GetJiraClient(cfg)
	assert.Nil(t, err)

	command := bot.Commands{}
	command.AddCommand(newCommentCommand(jiraClient, slackClient, cfg))

	t.Run("No match", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "quatsch"

		actual := command.Run(event)
		assert.Equal(t, false, actual)
	})

	t.Run("comment on not existing ticket", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "add comment to ticket NOPE-1234 thats true!"

		slackClient.On("ReplyError", event, mock.Anything)

		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})
}
