package mocks

import (
	"net/url"
	"testing"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func AssertSlackJSON(t *testing.T, slackClient *SlackClient, event slack.MessageEvent, expected url.Values) {
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
