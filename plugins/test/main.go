package main

//go:generate go build -o ./test.so ./main.go

import (
	"fmt"
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/client"
)

func GetCommands(cfg *bot.Bot, slack client.SlackClient) bot.Commands {
	fmt.Println("loading...")
	base := bot.BaseCommand{SlackClient: slack}

	commands := bot.Commands{}
	commands.AddCommand(testCommand{base})

	return commands
}

type testCommand struct {
	bot.BaseCommand
}

func (c testCommand) GetMatcher() matcher.Matcher {
	return matcher.NewTextMatcher("foo", func(match matcher.Result, message msg.Message) {
		c.SendMessage(message, "yep!")
	})
}

func main() {
	bot.ServePlugin(&bot.SlackBotPlugin{
		GetCommands: GetCommands,
	})
}
