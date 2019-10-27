package queue

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/storage"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/mocks"
	"github.com/nlopes/slack"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestQueue(t *testing.T) {
	client.InternalMessages = make(chan slack.MessageEvent, 2)
	slackClient := &mocks.SlackClient{}

	after := storage.MockStorage()
	defer after()

	event := slack.MessageEvent{}
	event.User = "testUser1"

	logger, _ := test.NewNullLogger()

	command := bot.Commands{}
	command.AddCommand(NewQueueCommand(slackClient, logger))
	command.AddCommand(NewListCommand(slackClient))

	t.Run("Invalid command", func(t *testing.T) {
		event.Text = "I have a queuedCommand"
		actual := command.Run(event)
		assert.Equal(t, false, actual)
		assert.Empty(t, client.InternalMessages)
	})

	t.Run("No command running", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "queue reply test"

		slackClient.On("ReplyError", event, fmt.Errorf("You have to call this command when another long running command is already running")).Return(true)
		actual := command.Run(event)
		assert.Equal(t, true, actual)
		assert.Empty(t, client.InternalMessages)
	})

	t.Run("No command from other user running", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "queue reply test"

		AddRunningCommand(slack.MessageEvent{
			Msg: slack.Msg{
				User: "testUser2",
			}},
			"",
		)

		slackClient.On("ReplyError", event, fmt.Errorf("You have to call this command when another long running command is already running")).Return(true)
		actual := command.Run(event)
		assert.Equal(t, true, actual)
		assert.Empty(t, client.InternalMessages)
	})

	t.Run("Test queue command", func(t *testing.T) {
		now := time.Now()
		event.Timestamp = strconv.Itoa(int(now.Unix()))
		done := AddRunningCommand(event, "test")

		msgRef := slack.NewRefToMessage(event.Channel, event.Timestamp)

		slackClient.On("AddReaction", waitIcon, msgRef)
		slackClient.On("AddReaction", doneIcon, msgRef)
		slackClient.On("RemoveReaction", waitIcon, msgRef)

		event.Text = "queue reply test"
		actual := command.Run(event)
		assert.Equal(t, true, actual)
		assert.Empty(t, client.InternalMessages)

		// list queue
		event.Text = "list queue"
		slackClient.On("Reply", event, mock.MatchedBy(func(input string) bool {
			return strings.HasPrefix(input, "1 queued commands")
		}))

		slackClient.On("GetReactions", msgRef, slack.NewGetReactionsParameters()).Return(
			[]slack.ItemReaction{
				{Name: "test"},
				{Name: "foo"},
			},
			nil,
		)
		actual = command.Run(event)
		assert.Equal(t, true, actual)

		// list queue for current channel
		event.Text = "list queue in channel"
		slackClient.On("Reply", event, mock.MatchedBy(func(reply string) bool {
			expected := regexp.MustCompile("1 queued commands\\n - <@> \\(.* ago\\): ```test``` :test::foo:")
			return expected.MatchString(reply)
		}))
		actual = command.Run(event)
		assert.Equal(t, true, actual)

		// list queue for other channel
		event.Text = "list queue in channel"
		event.Channel = "C1212121"
		slackClient.On("Reply", event, mock.MatchedBy(func(input string) bool {
			return input == "0 queued commands\n"
		}))
		actual = command.Run(event)
		assert.Equal(t, true, actual)

		done <- true
		time.Sleep(time.Millisecond * 300)

		assert.NotEmpty(t, client.InternalMessages)

		handledEvent := <-client.InternalMessages
		assert.Equal(t, handledEvent, slack.MessageEvent{
			Msg: slack.Msg{
				Timestamp: event.Timestamp,
				User:      "testUser1",
				Text:      "reply test",
			},
		})
	})
}
