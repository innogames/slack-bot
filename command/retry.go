package command

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/storage"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/innogames/slack-bot/client"
	"github.com/slack-go/slack"
)

const storageKey = "user_history"

var repeatRegexp = util.CompileRegexp("(retry|repeat)")

// NewRetryCommand store the history of the commands of the user sent to the bot in a local storage
// With "retry" the most recent command of the channel will be repeated
func NewRetryCommand(slackClient client.SlackClient) bot.Command {
	return &retryCommand{
		slackClient,
	}
}

type retryCommand struct {
	slackClient client.SlackClient
}

func (c *retryCommand) GetMatcher() matcher.Matcher {
	return matcher.WildcardMatcher(c.Execute)
}

func (c *retryCommand) Execute(event slack.MessageEvent) bool {
	if event.SubType == "internal" {
		return false
	}

	key := util.GetFullEventKey(event)
	shouldRetry := repeatRegexp.MatchString(event.Text)
	if !shouldRetry {
		storage.Write(storageKey, key, event.Text)
		return false
	}

	var lastCommand string
	storage.Read(storageKey, key, &lastCommand)
	if lastCommand != "" {
		c.slackClient.Reply(event, fmt.Sprintf("Executing command: %s", lastCommand))

		newMessage := event
		newMessage.Text = lastCommand
		client.InternalMessages <- msg.FromSlackEvent(newMessage)
	} else {
		c.slackClient.Reply(event, "Sorry, no history found.")
	}

	return true
}

func (c *retryCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "repeat",
			Description: "repeat the last executed command",
			Examples: []string{
				"retry",
				"repeat",
			},
		},
	}
}
