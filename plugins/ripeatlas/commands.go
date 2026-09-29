// Package ripeatlas is a slack-bot plugin to run queries against the RIPE Atlas API to debug network issues
package ripeatlas

import (
	"errors"

	"github.com/innogames/slack-bot/v2/bot"
)

var category = bot.Category{
	Name:        "RIPE Atlas",
	Description: "Run queries against the RIPE Atlas API to debug network issues",
}

func init() {
	bot.RegisterPlugin(bot.Plugin{
		Name:  "ripeatlas",
		Setup: setup,
	})
}

func setup(ctx *bot.PluginContext) (bot.Commands, error) {
	var commands bot.Commands

	cfg := defaultConfig
	if err := ctx.LoadConfig(&cfg); err != nil {
		return commands, err
	}
	if !cfg.IsEnabled() {
		return commands, errors.New(`no "api_key" defined in the config`)
	}

	base := ctx.BaseCommand()
	commands.AddCommand(
		&creditsCommand{base, cfg},
		&tracerouteCommand{base, cfg},
	)

	return commands, nil
}
