package mocks

import (
	"encoding/json"
	"fmt"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/innogames/slack-bot/client"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"strings"
	"testing"
	"time"
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

// AssertSlackBlocks test helper to assert a given JSON representation of "Blocks"
func AssertSlackBlocks(t *testing.T, slackClient *SlackClient, message msg.Ref, expectedJSON string) {
	t.Helper()

	slackClient.On("SendBlockMessage", message, mock.MatchedBy(func(givenBlocks []slack.Block) bool {
		givenJSON, err := json.Marshal(givenBlocks)
		assert.Nil(t, err)

		if expectedJSON != string(givenJSON) {
			fmt.Println(expectedJSON)
			fmt.Println("vs")
			fmt.Println(string(givenJSON))
		}

		return expectedJSON == string(givenJSON)
	})).Once().Return("")
}

// AssertContainsSlackBlocks is a small test helper to check for certain slack.Block
func AssertContainsSlackBlocks(t *testing.T, slackClient *SlackClient, message msg.Ref, block slack.Block) {
	t.Helper()

	slackClient.On("SendBlockMessage", message, mock.MatchedBy(func(givenBlocks []slack.Block) bool {
		givenJSON, err := json.Marshal(givenBlocks)
		assert.Nil(t, err)
		expectedJSONBlock, err := json.Marshal(block)
		assert.Nil(t, err)

		return strings.Contains(string(givenJSON), string(expectedJSONBlock))
	}), mock.Anything).Once().Return("")
}

// WaitTillHavingInternalMessage blocks until there is a internal message queued
func WaitTillHavingInternalMessage() {
	deadline := time.Now().Add(time.Second * 2)
	for {
		if len(client.InternalMessages) >= 1 {
			return
		}

		if time.Now().After(deadline) {
			log.Fatalf("No new internal message after 2 seconds!")
		}

		time.Sleep(time.Millisecond * 10)
	}
}
