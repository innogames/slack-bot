package admin

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/stats"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/innogames/slack-bot/v2/bot/version"
	"github.com/innogames/slack-bot/v2/command/queue"
)

// bots uptime
var startTime = time.Now()

// newStatsCommand shows a bunch of runtime statistics of the bot (admin-only)
func newStatsCommand(base bot.BaseCommand, cfg *config.Config) bot.Command {
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

	result.addNewSection("Overall stats")
	result.addValue("Total commands executed", formatStats(stats.TotalCommands))
	result.addValue("Unknown commands", formatStats(stats.UnknownCommands))
	result.addValue("Unauthorized commands", formatStats(stats.UnauthorizedCommands))
	result.addValue("Handled interactions/buttons", formatStats(stats.Interactions))

	result.addNewSection("Server runtime")
	result.addValue("Registered crons", util.FormatInt(len(c.cfg.Crons)))
	result.addValue("Queued commands", util.FormatInt(queue.CountCurrentJobs()))
	result.addValue("Goroutines", util.FormatInt(runtime.NumGoroutine()))
	result.addValue("Mem Alloc", util.FormatBytes(m.Alloc))
	result.addValue("Mem Sys", util.FormatBytes(m.Sys))
	result.addValue("Uptime", util.FormatDuration(time.Since(startTime)))
	result.addValue("NumGC (since start)", util.FormatInt(int(m.NumGC)))
	result.addValue("Bot Version", version.Version)
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

	return util.FormatInt(int(value))
}

func (c *statsCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "bot stats",
			Description: "display runtime stats from the bots, like total processed commands etc",
		},
	}
}
