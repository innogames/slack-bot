package admin

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/bot/stats"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/command/queue"
	"github.com/slack-go/slack"
	"runtime"
	"strings"
)

// NewStatsCommand shows a bunch of runtime statistics of the bot (admin-only)
func NewStatsCommand(slackClient client.SlackClient, cfg config.Config) bot.Command {
	return &statsCommand{slackClient, cfg}
}

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
	result := statsResult{}
	result.WriteString("Here are some current stats:\n")

	c.collectStats(&result)
	c.slackClient.Reply(event, result.String())
}

func (c statsCommand) collectStats(result *statsResult) {
	result.addNewSection("Runtime stats")
	result.addValue("Goroutines", fmt.Sprintf("%d goroutines", runtime.NumGoroutine()))
	// todo https://github.com/jtaczanowski/go-runtime-stats/blob/master/collector/collector.go

	result.addNewSection("Commands")
	result.addValue("Total Commands", formatStats(stats.TotalCommands))
	result.addValue("Unknown Commands", formatStats(stats.UnknownCommands))
	result.addValue("Unauthorized Commands", formatStats(stats.UnauthorizedCommands))
	result.addValue("Queued commands", fmt.Sprintf("%d", queue.CountCurrentJobs()))
}

type statsResult struct {
	strings.Builder
}

func (s *statsResult) addNewSection(section string) {
	s.WriteString(fmt.Sprintf("*%s*:\n", section))
}

func (s *statsResult) addValue(name string, value string) {
	s.WriteString(fmt.Sprintf("- %s: %s\n", name, value))
}

func formatStats(key string) string {
	value, err := stats.Get(key)
	if err != nil {
		return "unknown"
	}

	return fmt.Sprintf("%d", value)
}

func (c statsCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "bot stats",
			Description: "display runtime stats from the bots, like total processed commands etc",
		},
	}
}
