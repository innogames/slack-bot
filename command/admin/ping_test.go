package admin

import (
	"fmt"
	"testing"
	"time"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
)

func TestPing(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	botLog := newPingCommand(base)

	command := bot.Commands{}
	command.AddCommand(botLog)

	t.Run("test ping", func(t *testing.T) {
		// expect message to run for 1min
		msgTime := time.Now().Add(-time.Minute)

		message := msg.Message{}
		message.Text = "ping"
		message.Timestamp = fmt.Sprintf("%d.000000", msgTime.Unix())

		fmt.Println(message.Timestamp)
		mocks.AssertSlackMessageRegexp(slackClient, message, `^PONG in 1m`)

		actual := command.Run(message)
		assert.True(t, actual)
	})
}
