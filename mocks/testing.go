package mocks

import (
	"github.com/innogames/slack-bot/bot/msg"
	"net/url"
	"testing"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func AssertSlackMessage(slackClient *SlackClient, ref msg.Ref, text string) {
	slackClient.On("SendMessage", ref, text).Once().Return("")
}

// AssertSlackJSON is a test helper to assert full slack attachments
func AssertSlackJSON(t *testing.T, slackClient *SlackClient, message msg.Ref, expected url.Values) {
	slackClient.On("SendMessage", message, "", mock.MatchedBy(func(option slack.MsgOption) bool {
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
