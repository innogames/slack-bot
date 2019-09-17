package pullrequest

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/mocks"
	"github.com/nlopes/slack"
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
		event := slack.MessageEvent{}
		event.Text = "quatsch"

		actual := command.Run(event)
		assert.Equal(t, false, actual)
	})
	// todo add real test
}
