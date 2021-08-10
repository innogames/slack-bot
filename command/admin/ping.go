package admin

import (
	"fmt"
	"time"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/util"
)

// newPingCommand just prints a PING with the needed time from client->slack->bot server
func newPingCommand(base bot.BaseCommand) bot.Command {
	return &pingCommand{
		base,
	}
}

type pingCommand struct {
	bot.BaseCommand
}

func (c *pingCommand) GetMatcher() matcher.Matcher {
	return matcher.NewTextMatcher("ping", c.ping)
}

func (c *pingCommand) ping(match matcher.Result, message msg.Message) {
	c.SendMessage(message, fmt.Sprintf(
		"PONG in %s",
		util.FormatDuration(time.Since(message.GetTime())),
	))
}

func (c *pingCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "ping",
			Description: "just prints a PING with the needed time from client->slack->bot server",
			Examples: []string{
				"ping",
			},
		},
	}
}
