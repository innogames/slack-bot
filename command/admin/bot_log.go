package admin

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
)

const logChars = 4000

// newBotLogCommand prints the recent bot.log as slack command
func newBotLogCommand(base bot.BaseCommand, cfg *config.Config) bot.Command {
	return &botLogCommand{
		base,
		cfg,
	}
}

type botLogCommand struct {
	bot.BaseCommand
	cfg *config.Config
}

func (c *botLogCommand) GetMatcher() matcher.Matcher {
	return matcher.NewAdminMatcher(
		c.cfg.AdminUsers,
		c.SlackClient,
		matcher.NewTextMatcher("bot log", c.showBotLog),
	)
}

func (c *botLogCommand) showBotLog(match matcher.Result, message msg.Message) {
	log := c.readFile(c.cfg.Logger.File, logChars)
	parts := strings.SplitN(string(log), "\n", 2)
	if len(parts) <= 1 {
		c.SendMessage(message, "No logs so far")
		return
	}

	c.SendMessage(message, fmt.Sprintf("The most recent messages:\n```%s```", parts[1]))
}

func (c *botLogCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "bot log",
			Description: "displays the last log entries of the bot",
			Examples: []string{
				"bot log",
			},
		},
	}
}

// get the last X characters from the given file
func (c *botLogCommand) readFile(filename string, chars int64) []byte {
	buf := make([]byte, chars)
	file, err := os.Open(filename) // #nosec
	if err != nil {
		return buf
	}
	defer file.Close()

	stat, _ := os.Stat(filename)
	start := stat.Size() - chars
	if start < 0 {
		start = 0
	}

	file.ReadAt(buf, start)

	return bytes.Trim(buf, "\x00")
}
