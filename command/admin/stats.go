package admin

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/slack-go/slack"
	"runtime"
)

// NewStatsCommand shows a bunch of runtime statistics of the bot (admin-only)
func NewStatsCommand(slackClient client.SlackClient, cfg config.Config) bot.Command {
	return &statsCommand{slackClient, cfg}
}

type stats map[string]string

type statsCommand struct {
	slackClient client.SlackClient
	cfg         config.Config
}

func (c statsCommand) GetMatcher() matcher.Matcher {
	return matcher.NewAdminMatcher(
		c.cfg,
		c.slackClient,
		matcher.NewTextMatcher("bot stats", c.Stats),
	)
}

func (c statsCommand) Stats(match matcher.Result, event slack.MessageEvent) {
	currentStats := c.collectStats()

	message := "Here are some current stats:\n"
	for key, value := range currentStats {
		message += fmt.Sprintf("- %s: %s\n", key, value)
	}

	c.slackClient.Reply(event, message)
}

func (c statsCommand) collectStats() stats {
	currentStats := make(stats)

	// todo https://github.com/jtaczanowski/go-runtime-stats/blob/master/collector/collector.go
	currentStats["goroutines"] = fmt.Sprintf("%d goroutines", runtime.NumGoroutine())

	return currentStats
}

func (c statsCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "bot stats",
			Description: "display runtime stats from the bots, like total processed commands etc",
		},
	}
}
