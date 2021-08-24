package admin

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
)

func TestBotLog(t *testing.T) {
	testFile := "test.log"

	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	cfg := &config.Config{}
	cfg.Logger.File = testFile
	cfg.AdminUsers = []string{
		"UADMIN",
	}

	botLog := newBotLogCommand(base, cfg)

	command := bot.Commands{}
	command.AddCommand(botLog)

	t.Run("invalid command", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "log log log"

		actual := command.Run(message)
		assert.False(t, actual)
	})

	t.Run("display log without history", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "bot log"
		message.User = "UADMIN"

		mocks.AssertSlackMessage(slackClient, message, "No logs so far")

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("display log history", func(t *testing.T) {
		ioutil.WriteFile(testFile, []byte("test\nfoo\nbar"), 0o600)
		defer os.Remove(testFile)

		message := msg.Message{}
		message.Text = "bot log"
		message.User = "UADMIN"

		mocks.AssertSlackMessage(slackClient, message, "The most recent messages:\n```foo\nbar```")

		actual := command.Run(message)
		assert.True(t, actual)
	})
}
