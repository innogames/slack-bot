package pullrequest

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBitbucket(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	cfg := config.Config{}

	command := bot.Commands{}
	cmd := newBitbucketCommand(slackClient, cfg)
	command.AddCommand(cmd)

	t.Run("invalid command", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "quatsch"

		actual := command.Run(message)
		assert.Equal(t, false, actual)
	})
}
