package main

//go:generate go build -buildmode=plugin -o ../test.so ./test.go

import (
	"text/template"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/client"
)

func GetCommands(cfg *config.Config, slack client.SlackClient) bot.Commands {
	base := bot.BaseCommand{SlackClient: slack}

	commands := bot.Commands{}
	commands.AddCommand(testCommand{base})

	return commands
}

func GetTemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"test": func() string {
			return "it works"
		},
	}
}

type testCommand struct {
	bot.BaseCommand
}

func (c testCommand) GetMatcher() matcher.Matcher {
	return matcher.NewTextMatcher("foo", func(match matcher.Result, message msg.Message) {
		c.SendMessage(message, "yep!")
	})
}
