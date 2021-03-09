package admin

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/stats"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/innogames/slack-bot/command/queue"
)

// bots uptime
var startTime = time.Now()

// NewStatsCommand shows a bunch of runtime statistics of the bot (admin-only)
func NewStatsCommand(base bot.BaseCommand, cfg *config.Config) bot.Command {
	return &statsCommand{base, cfg}
}

type statsCommand struct {
	bot.BaseCommand
	cfg *config.Config
}

func (c *statsCommand) GetMatcher() matcher.Matcher {
	return matcher.NewAdminMatcher(
		c.cfg.AdminUsers,
		c.SlackClient,
		matcher.NewTextMatcher("bot stats", c.stats),
	)
}

func (c *statsCommand) stats(match matcher.Result, message msg.Message) {
	result := statsResult{}
	result.addLine("Here are some current stats:")

	c.collectStats(&result)
	c.SendMessage(message, result.String())
}

func (c *statsCommand) collectStats(result *statsResult) {
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)

	result.addNewSection("Processed commands")
	result.addValue("Total commands executed", formatStats(stats.TotalCommands))
	result.addValue("Unknown Commands", formatStats(stats.UnknownCommands))
	result.addValue("Unauthorized Commands", formatStats(stats.UnauthorizedCommands))
	result.addValue("Interactions", formatStats(stats.Interactions))
	result.addValue("Queued commands", fmt.Sprintf("%d", queue.CountCurrentJobs()))
	result.addValue("Crons", fmt.Sprintf("%d", len(c.cfg.Crons)))

	result.addNewSection("Server Runtime")
	result.addValue("Uptime", util.FormatDuration(time.Since(startTime)))
	result.addValue("Goroutines", fmt.Sprintf("%d", runtime.NumGoroutine()))
	result.addValue("Mem Alloc", util.FormatBytes(m.Alloc))
	result.addValue("Mem Sys", util.FormatBytes(m.Sys))
	result.addValue("NumGC", fmt.Sprintf("%d", m.NumGC))
	result.addValue("Bot Version", bot.Version)
	result.addValue("Go Version", runtime.Version())
}

type statsResult struct {
	strings.Builder
}

func (s *statsResult) addNewSection(section string) {
	s.addLine(fmt.Sprintf("*%s*:", section))
}

func (s *statsResult) addValue(name string, value string) {
	s.addLine(fmt.Sprintf("• %s: %s", name, value))
}

func (s *statsResult) addLine(line string) {
	_, _ = s.WriteString(line + "\n")
}

func formatStats(key string) string {
	value, err := stats.Get(key)
	if err != nil {
		return "0"
	}

	return strconv.Itoa(int(value))
}

func (c *statsCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "bot stats",
			Description: "display runtime stats from the bots, like total processed commands etc",
		},
	}
}
