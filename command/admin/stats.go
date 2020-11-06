package admin

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/slack-go/slack"
)

// NewDelayCommand delays the command execution by the given time
func NewStatsCommand(slackClient client.SlackClient, cfg config.Config) bot.Command {
	return &statsCommand{slackClient: slackClient}
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
	currentStats := collectStats()

	message := "Here are some current stats:\n"
	for key, value := range currentStats {
		message += fmt.Sprintf("- %s: %s", key, value)
	}

	c.slackClient.Reply(event, message)
}

func collectStats() stats {
	currentStats := make(stats)

	// todo https://github.com/jtaczanowski/go-runtime-stats/blob/master/collector/collector.go
	currentStats["test"] = "100 rzc"
	currentStats["bar"] = "1212 se/s"

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
