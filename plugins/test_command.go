package main

import (
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/innogames/slack-bot/v2/client"
)

type testCommand struct {
	bot.BaseCommand
}

func (c *testCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewPrefixMatcher("foo", c.reply),
	)
}

func (c *testCommand) reply(match matcher.Result, message msg.Message) {
	text := match.GetString(util.FullMatch)
	if text == "" {
		return
	}

	c.SendMessage(message, text)
}

func start(_ *bot.Bot, slackClient client.SlackClient) bot.Commands {
	commands := bot.Commands{}

	commands.AddCommand(&testCommand{bot.BaseCommand{SlackClient: slackClient}})

	return commands
}

//nolint:gochecknoinits // init required for plugin registration
func init() {
	bot.RegisterPlugin(bot.Plugin{
		Init: start,
	})
}
