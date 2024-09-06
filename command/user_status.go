package command

import (
	"fmt"
	"time"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/command/queue"
)

const notifyCheckInterval = time.Minute * 1

// Command which informs the user when the given user got active
func newUserStatusCommand(base bot.BaseCommand) *userStatus {
	return &userStatus{
		base,
		notifyCheckInterval,
	}
}

type userStatus struct {
	bot.BaseCommand
	checkInterval time.Duration
}

func (c *userStatus) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher(`notify user <@(?P<user>.*)> (?P<status>(away|active))`, c.NotifyUserActive)
}

func (c *userStatus) NotifyUserActive(match matcher.Result, message msg.Message) {
	user := match.GetString("user")
	expectedStatus := match.GetString("status")

	// in case of bot restart: restart this command again
	runningCommand := queue.AddRunningCommand(message, message.Text)

	c.AddReaction("⌛", message)
	go func() {
		defer c.RemoveReaction("⌛", message)
		defer runningCommand.Done()

		for {
			presence, err := c.SlackClient.GetUserPresence(user)
			if err != nil {
				c.ReplyError(message, err)
				return
			}

			if presence.Presence == expectedStatus {
				c.SendMessage(message, fmt.Sprintf("User <@%s> is %s now!", user, presence.Presence))
				return
			}

			time.Sleep(c.checkInterval)
		}
	}()
}

func (c *userStatus) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "notify user active",
			Description: "Inform you if the given user change the slack status to active.",
			Examples: []string{
				"notify user @myboss active",
			},
		},
	}
}
