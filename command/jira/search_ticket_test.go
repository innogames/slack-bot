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

func TestJira(t *testing.T) {
	slackClient := mocks.SlackClient{}

	// todo fake http client
	cfg := config.Jira{
		Host:    "https://issues.apache.org/jira/",
		Project: "ZOOKEEPER",
	}
	jiraClient, err := client.GetJiraClient(cfg)
	assert.Nil(t, err)

	command := bot.Commands{}
	command.AddCommand(NewJiraCommand(jiraClient, &slackClient, cfg))

	t.Run("No match", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "quatsch"

		actual := command.Run(event)
		assert.Equal(t, false, actual)
	})

	// todo: it just check for valid 200 but for test the result/attachment yet!
	t.Run("search existing ticket", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "jira ZOOKEEPER-3456"

		slackClient.On("SendMessage", event, "", mock.MatchedBy(func(input slack.MsgOption) bool {
			return true
		})).Return("")

		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})

	t.Run("search invalid ticket", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "jira ZOOKEEPER-10000000000"

		slackClient.On("Reply", event, "Issue Does Not Exist: Request failed. Please analyze the request body for more details. Status code: 404")

		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})

	t.Run("search existing ticket", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "jql FOO=BAR"

		slackClient.On("Reply", event, "Field 'FOO' does not exist or this field cannot be viewed by anonymous users.: Request failed. Please analyze the request body for more details. Status code: 400")
		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})
}

func TestConvertMarkdown(t *testing.T) {
	message := "h1. hallo how are {code}you{code}?"
	actual := convertMarkdown(message)

	assert.Equal(t, "hallo how are ```you```?", actual)
}

func BenchmarkConvertMarkdown(b *testing.B) {
	message := "h1. hallo how are {code}you{code}?"

	for i := 0; i < b.N; i++ {
		convertMarkdown(message)
	}
}

func BenchmarkConvertMarkdownNoMatch(b *testing.B) {
	message := "hallo how are you?"

	for i := 0; i < b.N; i++ {
		convertMarkdown(message)
	}
}
