package command

import (
	"net/url"
	"testing"

	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/mocks"
	"github.com/stretchr/testify/assert"
)

func TestAddLink(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	command := bot.Commands{}
	command.AddCommand(NewAddLinkCommand(base))

	t.Run("invalid command", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "add a link"

		actual := command.Run(message)
		assert.False(t, actual)
	})

	t.Run("add link", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "add link google <https://google.com>"

		expected := url.Values{}
		expected.Add("attachments", `[{"actions":[{"name":"","text":"google","style":"default","type":"button","url":"https://google.com"}],"blocks":null}]`)
		mocks.AssertSlackJSON(t, slackClient, message, expected)

		actual := command.Run(message)
		assert.True(t, actual)
	})
}
