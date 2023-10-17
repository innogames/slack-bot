package vcs

import (
	"fmt"
	"strings"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/innogames/slack-bot/v2/client/vcs"
	"golang.org/x/exp/slices"
)

func GetCommands(base bot.BaseCommand, config *config.Config) bot.Commands {
	var commands bot.Commands

	if config.BranchLookup.IsEnabled() {
		commands.AddCommand(&vcsCommand{
			base,
			config,
		})
	}

	return commands
}

// RunAsync registers regular branch updates in the background with proper stopping on exit
func (c *vcsCommand) RunAsync(ctx *util.ServerContext) {
	vcs.InitBranchWatcher(c.cfg, ctx)
}

type vcsCommand struct {
	bot.BaseCommand
	cfg *config.Config
}

func (c *vcsCommand) GetMatcher() matcher.Matcher {
	return matcher.NewTextMatcher("list branches", c.listBranches)
}

func (c *vcsCommand) listBranches(_ matcher.Result, message msg.Message) {
	branches := vcs.GetBranches()
	slices.Sort(branches)

	response := strings.Builder{}
	response.WriteString(fmt.Sprintf("Found %d branches:\n", len(branches)))
	for _, branch := range branches {
		response.WriteString(fmt.Sprintf("- %s\n", branch))
	}

	c.SendMessage(message, response.String())
}

func (c *vcsCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "list branch",
			Description: "List all found VCS branches",
			Examples: []string{
				"list branches",
			},
		},
	}
}
