package command

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/storage"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/innogames/slack-bot/client"
)

const storageKey = "user_history"

var repeatRegexp = util.CompileRegexp("(retry|repeat)")

// NewRetryCommand store the history of the commands of the user sent to the bot in a local storage
// With "retry" the most recent command of the channel will be repeated
func NewRetryCommand(base bot.BaseCommand) bot.Command {
	return &retryCommand{
		base,
	}
}

type retryCommand struct {
	bot.BaseCommand
}

func (c *retryCommand) GetMatcher() matcher.Matcher {
	return matcher.WildcardMatcher(c.Execute)
}

func (c *retryCommand) Execute(ref msg.Ref, text string) bool {
	if ref.IsInternalMessage() {
		return false
	}

	key := ref.GetUniqueKey()
	shouldRetry := repeatRegexp.MatchString(text)
	if !shouldRetry {
		storage.Write(storageKey, key, text)
		return false
	}

	var lastCommand string
	storage.Read(storageKey, key, &lastCommand)
	if lastCommand != "" {
		c.SendMessage(ref, fmt.Sprintf("Executing command: %s", lastCommand))

		client.InternalMessages <- ref.WithText(lastCommand)
	} else {
		c.SendMessage(ref, "Sorry, no history found.")
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
