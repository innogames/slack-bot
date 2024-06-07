package main

//go:generate go build -o ./test.so ./main.go

import (
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
)

// Plugin implementation structure
type PluginImpl struct{}

func (p *PluginImpl) GetCommands() string {
	//	fmt.Println("loading...")
	//	base := bot.BaseCommand{SlackClient: slack}
	//
	//	commands := bot.Commands{}
	//	commands.AddCommand(testCommand{base})
	//
	//	return commands
	return "jop"
}

type testCommand struct {
	bot.BaseCommand
}

func (c testCommand) GetMatcher() matcher.Matcher {
	return matcher.NewTextMatcher("foo", func(match matcher.Result, message msg.Message) {
		c.SendMessage(message, "yep!")
	})
}

// Main function to serve the plugin
func main() {
	bot.ServePlugin(&PluginImpl{})
}
