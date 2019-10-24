package mocks

import (
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/url"
	"testing"
)

func AssertSlackJson(t *testing.T, slackClient *SlackClient, event slack.MessageEvent, expected url.Values) {
	slackClient.On("SendMessage", event, "", mock.MatchedBy(func(option slack.MsgOption) bool {
		_, values, _ := slack.UnsafeApplyMsgOptions(
			"token",
			"channel",
			"apiUrl",
			option,
		)

		expected.Add("token", "token")
		expected.Add("channel", "channel")

		assert.Equal(t, expected, values)

		return true
	})).Return("")
}
