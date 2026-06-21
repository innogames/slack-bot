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
	var text strings.Builder
	fmt.Fprintf(&text, "*%d crons:*\n", len(c.cfg))

	now := time.Now()
	for _, entry := range c.cron.Entries() {
		cfg, ok := c.entryToCfg[entry.ID]
		if !ok {
			continue
		}
		last := ""
		if !entry.Prev.IsZero() {
			last = fmt.Sprintf("last %s, ", entry.Prev)
		}
		fmt.Fprintf(&text,
			" - `%s`, %snext in %s (`%s`)\n",
			cfg.Schedule,
			last,
			util.FormatDuration(entry.Next.Sub(now)),
			strings.Join(cfg.Commands, "; "),
		)
	}

	c.SendMessage(message, text.String())
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
