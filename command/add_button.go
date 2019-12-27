package command

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/nlopes/slack"
)

// NewAddButtonCommand is more or less internal command to add a link button to the posted message
func NewAddButtonCommand(slackClient client.SlackClient) bot.Command {
	return &addButtonCommand{slackClient}
}

type addButtonCommand struct {
	slackClient client.SlackClient
}

func (c *addButtonCommand) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher("add button \"(?P<name>.*)\" \"(?P<command>.*)\"", c.AddLink)
}

func (c *addButtonCommand) AddLink(match matcher.Result, event slack.MessageEvent) {
	name := match.GetString("name")
	command := match.GetString("command")

	blocks := []slack.Block{
		client.GetInteraction(event, name, command),
	}

	c.slackClient.SendMessage(event, "", slack.MsgOptionBlocks(blocks...))
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
