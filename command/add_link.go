package command

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/client"
	"github.com/slack-go/slack"
)

// NewAddLinkCommand is more or less internal command to add a link button to the posted message
func NewAddLinkCommand(base bot.BaseCommand) bot.Command {
	return &addLinkCommand{base}
}

type addLinkCommand struct {
	bot.BaseCommand
}

func (c *addLinkCommand) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher("add link (?P<name>.*) <(?P<link>https:.*)>", c.AddLink)
}

func (c *addLinkCommand) AddLink(match matcher.Result, message msg.Message) {
	name := match.GetString("name")
	link := match.GetString("link")

	attachment := slack.Attachment{
		Actions: []slack.AttachmentAction{
			client.GetSlackLink(name, link),
		},
	}

	c.SendMessage(message, "", slack.MsgOptionAttachments(attachment))
}

func (c *addLinkCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "add link",
			Description: "adds a link button to the message",
			Examples: []string{
				"add link Google https://google.com",
				"add link Review https://stash.example.com/pr/12/review",
			},
		},
	}
}
