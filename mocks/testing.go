package mocks

import (
	"encoding/json"
	"fmt"
	"net/url"
	"testing"

	"github.com/innogames/slack-bot/bot/msg"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func AssertSlackMessage(slackClient *SlackClient, ref msg.Ref, text string) {
	slackClient.On("SendMessage", ref, text).Once().Return("")
}

// AssertSlackJSON is a test helper to assert full slack attachments
func AssertSlackJSON(t *testing.T, slackClient *SlackClient, message msg.Ref, expected url.Values) {
	t.Helper()

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
	})).Once().Return("")
}

// AssertSlackBlocks test helper to assert a given JSON representation of "Blocks"
func AssertSlackBlocks(t *testing.T, slackClient *SlackClient, message msg.Ref, expectedJSON string) {
	t.Helper()

	slackClient.On("SendBlockMessage", message, mock.MatchedBy(func(givenBlocks []slack.Block) bool {
		// replace the random tokens to fixed ones for easier mocking
		for i := range givenBlocks {
			if actionBlock, ok := givenBlocks[i].(*slack.ActionBlock); ok {
				if button, ok := actionBlock.Elements.ElementSet[0].(*slack.ButtonBlockElement); ok {
					button.Value = fmt.Sprintf("token-%d", i)
				}
			}
		}
		givenJSON, err := json.Marshal(givenBlocks)
		assert.Nil(t, err)

		fmt.Println(string(givenJSON))

		return expectedJSON == string(givenJSON)
	})).Once().Return("")
}
