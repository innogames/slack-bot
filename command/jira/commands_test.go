package jira

import (
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCommands(t *testing.T) {
	slackClient := &mocks.SlackClient{}

	cfg := &config.Jira{
		Host:    "https://issues.apache.org/jira/",
		Project: "ZOOKEEPER",
	}

	commands := GetCommands(cfg, slackClient)
	assert.Equal(t, 3, commands.Count())
}
