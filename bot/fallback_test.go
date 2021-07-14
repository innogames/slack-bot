package bot

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

type testCommand struct{}

func (c testCommand) GetHelp() []Help {
	return []Help{
		{
			Command:     "delay",
			Description: "delay a command by the given offset",
			Category: Category{
				Name: "testCategory",
			},
			Examples: []string{
				"delay 1h rely remind me to go to toilet",
				"delay 15m30s trigger job DeployBeta",
				"delay 15min trigger job DeployBeta",
			},
		},
		{
			Command:     "stop delay",
			Description: "cancel a planned delayCommand",
			Examples: []string{
				"stop delay 1243",
			},
		},
	}
}

func (c testCommand) GetMatcher() matcher.Matcher {
	return matcher.NewVoidMatcher()
}

func TestFallback(t *testing.T) {
	slackClient := &client.Slack{}
	bot := &Bot{
		auth: &slack.AuthTestResponse{
			User: "test",
		},
		slackClient: slackClient,
	}
	commands := &Commands{}

	commands.AddCommand(
		testCommand{},
	)
	bot.commands = commands

	t.Run("sendFallbackMessage", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "delay"

		bot.sendFallbackMessage(message)

		assert.Len(t, client.InternalMessages, 1)
		actual := <-client.InternalMessages

		assert.Equal(t, "help delay", actual.GetText())
	})

	t.Run("getBestMatchingHelp", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "delay"

		actual := getBestMatchingHelp(bot, "reply")
		assert.Equal(t, "delay", actual.Command)
	})

	t.Run("getBestMatchingHelpWithoutAlternative", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "djasiodsadUFBUẞFif"

		actual := getBestMatchingHelp(bot, "reply")
		assert.Equal(t, "delay", actual.Command)
	})
}
