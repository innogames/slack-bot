package command

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/client"
	"github.com/slack-go/slack"
)

// NewAddButtonCommand is more or less internal command to add a link button to the posted message
func NewAddButtonCommand(base bot.BaseCommand, cfg config.Slack) bot.Command {
	return &addButtonCommand{base, cfg}
}

type addButtonCommand struct {
	bot.BaseCommand
	cfg config.Slack
}

func (c *addButtonCommand) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher(`add button "(?P<name>.*)" "(?P<command>.*)"`, c.AddLink)
}

func (c *addButtonCommand) AddLink(match matcher.Result, message msg.Message) {
	name := match.GetString("name")
	command := match.GetString("command")

	blocks := []slack.Block{
		slack.NewActionBlock("", client.GetInteractionButton(message, name, command)),
	}

	c.SendMessage(message, "", slack.MsgOptionBlocks(blocks...))
}

// IsEnabled checks if the http server is enabled to receive slack interactions
func (c *addButtonCommand) IsEnabled() bool {
	return c.cfg.CanHandleInteractions()
}

func (c *addButtonCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "add button",
			Description: "adds a button to the message which then performs any command",
			Examples: []string{
				"add button \"Start job\" \"trigger job XYZ\"",
			},
		},
	}
}
