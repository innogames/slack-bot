package main

import (
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"text/template"
)

func GetTemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"test": func() string {
			return "it works"
		},
	}
}

func GetCommands(cfg config.Config) bot.Commands {
	commands := bot.Commands{}
	commands.AddCommand(testCommand{})

	return commands
}

type testCommand struct {
}

func (c testCommand) GetMatcher() matcher.Matcher {
	return matcher.NewTextMatcher("foo", func(match matcher.Result, message msg.Message) {
	})
}
