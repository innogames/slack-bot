package cron

import (
	"strings"
	"testing"

	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCron(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	crons := []config.Cron{
		{
			Channel:  "#dev",
			Schedule: "0 0 * * *",
			Commands: []string{
				"reply foo",
				"reply bar",
			},
		},
		{
			Channel:  "#foo",
			Schedule: "0 0 * * *",
			Commands: []string{
				"{{.Name}}",
				"{{}}",
			},
		},
		{
			Channel:  "#invalid",
			Schedule: "0 ",
			Commands: []string{
				"invalid schedule",
			},
		},
	}
	command := NewCronCommand(base, crons).(*command)
	commands := bot.Commands{}
	commands.AddCommand(command)

	t.Run("List crons", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "list crons"
		slackClient.On("SendMessage", message, mock.MatchedBy(func(input string) bool {
			return strings.HasPrefix(input, "*3 crons:*\n - `0 0 * * *`, next in")
		})).Return("")
		actual := commands.Run(message)
		assert.True(t, actual)
	})

	t.Run("Execute", func(t *testing.T) {
		jobs := command.cron.Entries()
		assert.Len(t, jobs, 2)
		for _, job := range jobs {
			job.Job.Run()
		}

		baseMessage := msg.Message{}
		baseMessage.User = "cron"

		mocks.AssertQueuedMessage(t, baseMessage.WithText("reply foo"))
		mocks.AssertQueuedMessage(t, baseMessage.WithText("reply bar"))
	})

	t.Run("Test help", func(t *testing.T) {
		help := commands.GetHelp()
		assert.NotNil(t, help)
	})
}
