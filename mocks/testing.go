package mocks

import (
	"encoding/json"
	"fmt"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/innogames/slack-bot/client"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"strings"
	"testing"
)

func AssertSlackMessage(slackClient *SlackClient, ref msg.Ref, text string) {
	slackClient.On("SendMessage", ref, text).Once().Return("")
}

func AssertReaction(slackClient *SlackClient, reaction string, ref msg.Ref) {
	slackClient.On("AddReaction", util.Reaction(reaction), ref).Once()
}

func AssertRemoveReaction(slackClient *SlackClient, reaction string, ref msg.Ref) {
	slackClient.On("RemoveReaction", util.Reaction(reaction), ref).Once()
}

func AssertError(slackClient *SlackClient, ref msg.Ref, errorIn interface{}) {
	var err error
	switch e := errorIn.(type) {
	case string:
		err = fmt.Errorf(e)
	case error:
		err = e
	}
	slackClient.On("ReplyError", ref, err).Once()
}

func AssertQueuedMessage(t *testing.T, expected msg.Message) {
	t.Helper()

	actual := <-client.InternalMessages
	assert.Equal(t, actual, expected)
}

// AssertSlackJSON is a test helper to assert full slack attachments
func AssertSlackJSON(t *testing.T, slackClient *SlackClient, message msg.Ref, expected string) {
	t.Helper()

	slackClient.On("SendMessage", message, "", mock.MatchedBy(func(option slack.MsgOption) bool {
		_, values, _ := slack.UnsafeApplyMsgOptions(
			"token",
			"channel",
			"apiUrl",
			option,
		)

		assert.Equal(t, expected, values.Get("attachments"))

		return true
	})).Once().Return("")
}

// AssertSlackJSONContains is a test helper to assert parts of the JSON slack attachments
func AssertSlackJSONContains(t *testing.T, slackClient *SlackClient, message msg.Ref, expected string) {
	t.Helper()

	slackClient.On("SendMessage", message, "", mock.MatchedBy(func(option slack.MsgOption) bool {
		_, values, _ := slack.UnsafeApplyMsgOptions(
			"token",
			"channel",
			"apiUrl",
			option,
		)

		return strings.Contains(values.Get("attachments"), expected)
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
