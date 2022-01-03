package mocks

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/innogames/slack-bot/v2/client"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// testLock is only used in integration test to avoid running some tests in parallel as we use one shared Message list
var testLock sync.Mutex

// AssertSlackMessage is a test helper to check for a given slack message
func AssertSlackMessage(slackClient *SlackClient, ref msg.Ref, text string) {
	slackClient.On("SendMessage", ref, text).Once().Return("")
}

// AssertSlackMessageRegexp is a test helper to check for a given slack message based on a regular expression
func AssertSlackMessageRegexp(slackClient *SlackClient, ref msg.Ref, pattern string) {
	slackClient.On("SendMessage", ref, mock.MatchedBy(func(text string) bool {
		re := regexp.MustCompile(pattern)
		return re.MatchString(text)
	})).Once().Return("")
}

// AssertReaction is a test helper to expect a given slack reaction to be added
func AssertReaction(slackClient *SlackClient, reaction string, ref msg.Ref) {
	slackClient.On("AddReaction", mock.MatchedBy(func(actualReaction util.Reaction) bool {
		return util.Reaction(reaction).ToSlackReaction() == actualReaction.ToSlackReaction()
	}), ref).Once()
}

// AssertRemoveReaction is a test helper to expect a given slack reaction to be removed
func AssertRemoveReaction(slackClient *SlackClient, reaction string, ref msg.Ref) {
	slackClient.On("RemoveReaction", util.Reaction(reaction), ref).Once()
}

// AssertError is a test helper which check for calls to "ReplyError"
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

// AssertQueuedMessage checks if a given internal message was queued
func AssertQueuedMessage(t *testing.T, expected msg.Message) {
	t.Helper()

	actual := <-client.InternalMessages
	assert.Equal(t, actual, expected)
}

// WaitForQueuedMessages waits until all "count" messages are queued and returns them in the returned function.
// We get a failed test when there are more messages than expected or not enough messages after 1 second of timeout.
func WaitForQueuedMessages(t *testing.T, count int) func() []msg.Message {
	t.Helper()

	wg := &sync.WaitGroup{}
	wg.Add(count)

	assert.Empty(t, client.InternalMessages)

	messages := make([]msg.Message, 0, count)

	go func() {
		for message := range client.InternalMessages {
			// mark as handled -> the command can add the next command
			// just some time to test if the concurrency works
			time.Sleep(time.Millisecond * 5)

			message.Done.Done()
			message.Done = nil
			messages = append(messages, message)
			wg.Done()
		}
	}()

	return func() []msg.Message {
		defer close(client.InternalMessages)

		timer := time.NewTimer(time.Second)
		go func() {
			<-timer.C
			t.Fail()
		}()

		wg.Wait()

		timer.Stop()

		assert.Len(t, messages, count)

		return messages
	}
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

		return expected == values.Get("attachments")
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

		if strings.Contains(string(givenJSON), string(expectedJSONBlock)) {
			// all good!
			return true
		}

		fmt.Println("not matching:", string(givenJSON), string(expectedJSONBlock))

		return false
	}), mock.Anything).Once().Return("")
}

// LockInternalMessages uses mutex to block other tests dealing with the central message queue
func LockInternalMessages() *sync.Mutex {
	testLock.Lock()

	// empty the queue in case there was previous leftover
	close(client.InternalMessages)
	client.InternalMessages = make(chan msg.Message, 10)

	return &testLock
}

// WaitTillHavingInternalMessage blocks until there is a internal message queued. If there is no message after 2s -> exit!
func WaitTillHavingInternalMessage() msg.Message {
	deadline := time.Second * 2
	timeout := time.NewTimer(deadline)

	select {
	case <-timeout.C:
		log.Fatalf("No new internal message after %s!", deadline)
		return msg.Message{}
	case message := <-client.InternalMessages:

		return message
	}
}
