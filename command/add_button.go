package command

import (
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/slack-go/slack"
)

// NewAddButtonCommand is more or less internal command to add a link button to the posted message
func NewAddButtonCommand(base bot.BaseCommand) bot.Command {
	return &addButtonCommand{base}
}

type addButtonCommand struct {
	bot.BaseCommand
}

func (c *addButtonCommand) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher(`add button "(?P<name>.*)" "(?P<command>.*)"`, c.addLink)
}

func (c *addButtonCommand) addLink(match matcher.Result, message msg.Message) {
	name := match.GetString("name")
	command := match.GetString("command")

	blocks := []slack.Block{
		slack.NewActionBlock("", client.GetInteractionButton(name, command)),
	}

	c.SendBlockMessage(message, blocks)
}

// IsEnabled checks if the http server is enabled to receive slack interactions
func (c *addButtonCommand) IsEnabled() bool {
	return c.CanHandleInteractions()
}

func (c *addButtonCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     `add button "<text>" "<command>"`,
			Category:    helperCategory,
			Description: "adds a button to the message which then performs any command",
			Examples: []string{
				"add button \"Start job\" \"trigger job XYZ\"",
			},
		},
	}
}
