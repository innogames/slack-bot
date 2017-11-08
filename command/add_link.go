package command

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/nlopes/slack"
)

// NewAddLinkCommand is more or less internal command to add a link button to the posted message
func NewAddLinkCommand(slackClient client.SlackClient) bot.Command {
	return &addLinkCommand{slackClient}
}

type addLinkCommand struct {
	slackClient client.SlackClient
}

func (c *addLinkCommand) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher("add link (?P<name>.*) <(?P<link>https:.*)>", c.AddLink)
}

func (c *addLinkCommand) AddLink(match matcher.Result, event slack.MessageEvent) {
	name := match.GetString("name")
	link := match.GetString("link")

	attachment := slack.Attachment{
		Actions: []slack.AttachmentAction{
			client.GetSlackLink(name, link),
		},
	}

	c.slackClient.SendMessage(event, "", slack.MsgOptionAttachments(attachment))
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
