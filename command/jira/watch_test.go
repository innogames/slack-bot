package jira

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/mocks"
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWatchJira(t *testing.T) {
	slackClient := &mocks.SlackClient{}

	// todo fake http client
	cfg := config.Jira{
		Host: "https://issues.apache.org/jira/",
	}
	jiraClient, err := client.GetJiraClient(cfg)
	assert.Nil(t, err)

	command := bot.Commands{}
	command.AddCommand(newWatchCommand(jiraClient, slackClient, cfg))

	t.Run("No match", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "quatsch"

		actual := command.Run(event)
		assert.Equal(t, false, actual)
	})

	t.Run("watch not existing ticket", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "watch ticket ZOOKEEPER-345600010"

		slackClient.On("Reply", event, "Issue Does Not Exist: Request failed. Please analyze the request body for more details. Status code: 404")

		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})
}
