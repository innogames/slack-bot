package cron

import (
	"fmt"
	"strings"
	"time"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/util"
)

func (c *command) GetMatcher() matcher.Matcher {
	return matcher.NewTextMatcher("list crons", c.listCrons)
}

func (c *command) listCrons(_ matcher.Result, message msg.Message) {
	text := fmt.Sprintf("*%d crons:*\n", len(c.cfg))

	now := time.Now()
	for i, entry := range c.cron.Entries() {
		last := ""
		if !entry.Prev.IsZero() {
			last = fmt.Sprintf("last %s, ", entry.Prev)
		}
		text += fmt.Sprintf(
			" - `%s`, %snext in %s (`%s`)\n",
			c.cfg[i].Schedule,
			last,
			util.FormatDuration(entry.Next.Sub(now)),
			strings.Join(c.cfg[i].Commands, "; "),
		)
	}

	c.SendMessage(message, text)
}

func (c *command) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "list crons",
			Description: "list the registered crons and the next execution time",
			HelpURL:     "https://github.com/innogames/slack-bot#cron",
			Examples: []string{
				"list crons",
			},
		},
	}
}
