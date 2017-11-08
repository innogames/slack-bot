package jira

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/config"
	"github.com/innogames/slack-bot/mocks"
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
	"gopkg.in/andygrunwald/go-jira.v1"
	"testing"
)

func TestJira(t *testing.T) {
	slackClient := mocks.SlackClient{}
	jiraClient := &jira.Client{}
	cfg := &config.Jira{}

	command := bot.Commands{}
	command.AddCommand(NewJiraCommand(jiraClient, &slackClient, *cfg))

	event := slack.MessageEvent{}
	event.Text = "quatsch"

	actual := command.Run(event)
	assert.Equal(t, false, actual)

	// todo add real test
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
