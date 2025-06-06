package cron

import (
	"strings"
	"testing"
	"time"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCron(t *testing.T) {
	slackClient := mocks.NewSlackClient(t)
	base := bot.BaseCommand{SlackClient: slackClient}

	crons := []config.Cron{
		{
			Channel:  "#dev",
			Schedule: "0 0 * * *",
			Commands: []string{
				"reply foo",
				`
					reply bar1
					reply bar2
				`,
			},
		},
		{
			Channel:  "#foo",
			Schedule: "0 0 * * *",
			Commands: []string{
				"{{.Name}}",
				"{{Name}}",
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

	t.Run("Run in background", func(t *testing.T) {
		ctx := util.NewServerContext()
		go command.RunAsync(ctx)
		time.Sleep(time.Millisecond * 10)
		ctx.StopTheWorld()
	})

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
		getMessages := mocks.WaitForQueuedMessages(t, 4)

		for _, job := range jobs {
			job.Job.Run()
		}

		baseMessage := msg.Message{}
		baseMessage.User = "cron"
		baseMessage.Channel = "#dev"

		actualMessages := getMessages()
		assert.Equal(t, baseMessage.WithText("reply foo"), actualMessages[0])
		assert.Equal(t, baseMessage.WithText("reply bar1"), actualMessages[1])
		assert.Equal(t, baseMessage.WithText("reply bar2"), actualMessages[2])
	})

	t.Run("Test help", func(t *testing.T) {
		help := commands.GetHelp()
		assert.NotNil(t, help)
	})
}
