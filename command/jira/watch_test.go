package jira

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
)

func TestWatchJira(t *testing.T) {
	slackClient := &mocks.SlackClient{}

	// todo fake http client
	cfg := &config.Jira{
		Host: "https://issues.apache.org/jira/",
	}
	jiraClient, err := client.GetJiraClient(cfg)
	assert.Nil(t, err)

	command := bot.Commands{}
	command.AddCommand(newWatchCommand(jiraClient, slackClient, cfg))

	t.Run("No match", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "quatsch"

		actual := command.Run(message)
		assert.False(t, actual)
	})

	t.Run("watch not existing ticket", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "watch ticket ZOOKEEPER-345600010"

		slackClient.On("SendMessage", message, "Issue Does Not Exist: request failed. Please analyze the request body for more details. Status code: 404").Return("")

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Test help", func(t *testing.T) {
		help := command.GetHelp()
		assert.Equal(t, 1, len(help))
	})
}
