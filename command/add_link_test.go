package command

import (
	"net/url"
	"testing"

	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/mocks"
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
)

func TestAddLink(t *testing.T) {
	slackClient := &mocks.SlackClient{}

	command := bot.Commands{}
	command.AddCommand(NewAddLinkCommand(slackClient))

	t.Run("invalid command", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "add a link"

		actual := command.Run(event)
		assert.Equal(t, false, actual)
	})

	t.Run("add link", func(t *testing.T) {
		event := slack.MessageEvent{}

		event.Text = "add link google <https://google.com>"

		expected := url.Values{}
		expected.Add("attachments", "[{\"actions\":[{\"name\":\"\",\"text\":\"google\",\"style\":\"default\",\"type\":\"button\",\"url\":\"https://google.com\"}]}]")
		mocks.AssertSlackJson(t, slackClient, event, expected)

		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})
}
