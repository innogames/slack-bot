package admin

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/bot/stats"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/command/queue"
	"github.com/slack-go/slack"
	"runtime"
	"strings"
	"time"
)

// bots uptime
var startTime = time.Now()

// NewStatsCommand shows a bunch of runtime statistics of the bot (admin-only)
func NewStatsCommand(slackClient client.SlackClient, cfg config.Config) bot.Command {
	return statsCommand{slackClient, cfg}
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
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)

	result.addNewSection("Processed commands")
	result.addValue("Total commands executed", formatStats(stats.TotalCommands))
	result.addValue("Unknown Commands", formatStats(stats.UnknownCommands))
	result.addValue("Unauthorized Commands", formatStats(stats.UnauthorizedCommands))
	result.addValue("Queued commands", fmt.Sprintf("%d", queue.CountCurrentJobs()))

	result.addNewSection("Server Runtime")
	result.addValue("Uptime", util.FormatDuration(time.Since(startTime)))
	result.addValue("Goroutines", fmt.Sprintf("%d", runtime.NumGoroutine()))
	result.addValue("Mem Alloc", util.FormatBytes(m.Alloc))
	result.addValue("Mem TotalAlloc", util.FormatBytes(m.TotalAlloc))
	result.addValue("Mem Sys", util.FormatBytes(m.Sys))
	result.addValue("NumGC", fmt.Sprintf("%d", m.NumGC))
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
		return "0"
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
