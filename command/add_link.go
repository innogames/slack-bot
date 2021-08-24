package command

import (
	"strings"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/client"
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
	return matcher.NewRegexpMatcher(`add link (?P<name>.*) <?(?P<link>https:.*)>?`, c.addLink)
}

func (c *addLinkCommand) addLink(match matcher.Result, message msg.Message) {
	name := match.GetString("name")
	link := match.GetString("link")

	attachment := slack.Attachment{
		Actions: []slack.AttachmentAction{
			client.GetSlackLink(
				strings.Trim(name, "'\""),
				strings.Trim(link, "<>"),
			),
		},
	}

	c.SendMessage(message, "", slack.MsgOptionAttachments(attachment))
}

func (c *addLinkCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "add link <text> <link>",
			Description: "adds a link button to the message",
			Category:    helperCategory,
			Examples: []string{
				"add link Google https://google.com",
				"add link 'Review this' https://stash.example.com/pr/12/review",
			},
		},
	}
}
