package cron

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/nlopes/slack"
	"strings"
	"time"
)

func (c *command) GetMatcher() matcher.Matcher {
	return matcher.NewTextMatcher("list crons", c.ListCrons)
}

func (c *command) ListCrons(match matcher.Result, event slack.MessageEvent) {
	message := fmt.Sprintf("*%d crons:*\n", len(c.cfg))

	now := time.Now()
	for i, entry := range c.cron.Entries() {
		message += fmt.Sprintf(
			" - `%s`, next in %s (`%s`)\n",
			c.cfg[i].Schedule,
			util.FormatDuration(entry.Next.Sub(now)),
			strings.Join(c.cfg[i].Commands, "; "),
		)
	}

	c.slackClient.Reply(event, message)
}

func (c *command) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "list crons",
			Description: "list the registered crons and the next execution time",
			Examples: []string{
				"list crons",
			},
		},
	}
}
