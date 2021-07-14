package jira

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
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
