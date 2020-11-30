package queue

import (
	"fmt"
	"github.com/innogames/slack-bot/bot/msg"
	"time"

	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/bot/storage"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/innogames/slack-bot/client"
	"github.com/slack-go/slack"
)

type listCommand struct {
	slackClient client.SlackClient
}

type filterFunc func(msg.Message) bool

// NewListCommand prints the list of all queued commands (blocking commands like running Jenkins jobs)
func NewListCommand(slackClient client.SlackClient) bot.Command {
	return &listCommand{
		slackClient,
	}
}

func (c *listCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewTextMatcher("list queue", c.ListAll),
		matcher.NewTextMatcher("list queue in channel", c.ListChannel),
	)
}

func (c *listCommand) ListAll(match matcher.Result, message msg.Message) {
	c.listQueue(message, func(event msg.Message) bool {
		return true
	})
}

func (c *listCommand) ListChannel(match matcher.Result, message msg.Message) {
	c.listQueue(message, func(queuedEvent msg.Message) bool {
		return message.GetChannel() == queuedEvent.GetChannel()
	})
}

func (c *listCommand) listQueue(message msg.Message, filter filterFunc) {
	keys, _ := storage.GetKeys(storageKey)
	now := time.Now()

	count := 0
	attachments := make([]slack.Attachment, 0, len(keys))

	var queuedEvent msg.Message
	for _, key := range keys {
		if err := storage.Read(storageKey, key, &queuedEvent); err != nil {
			continue
		}

		if !filter(queuedEvent) {
			continue
		}

		count++
		userID, _ := client.GetUser(queuedEvent.User)

		messageTime := queuedEvent.GetTime()
		timeAgo := now.Sub(messageTime)
		color := getColor(timeAgo)
		text := fmt.Sprintf(
			"<@%s> (%s, %s ago): ```%s``` %s \n",
			userID,
			messageTime.Format(time.Stamp),
			util.FormatDuration(timeAgo),
			queuedEvent.GetText(),
			c.getReactions(queuedEvent),
		)
		attachments = append(attachments, slack.Attachment{
			Text:  text,
			Color: color,
			MarkdownIn: []string{
				"text",
			},
		})
	}

	response := fmt.Sprintf("%d queued commands", count)

	c.slackClient.SendMessage(message, response, slack.MsgOptionAttachments(attachments...))
}

// get attachment color for a given message time
// older messages will be marked as red to ma them as more important
func getColor(timeAgo time.Duration) string {
	var color string
	if timeAgo.Hours() >= 24 {
		color = "#CC0000"
	} else {
		color = "#E0E000"
	}

	return color
}

func (c *listCommand) getReactions(ref msg.Ref) string {
	formattedReactions := ""
	msgRef := slack.NewRefToMessage(ref.GetChannel(), ref.GetTimestamp())
	reactions, _ := c.slackClient.GetReactions(msgRef, slack.NewGetReactionsParameters())

	for _, reaction := range reactions {
		formattedReactions += ":" + reaction.Name + ":"
	}
	return formattedReactions
}

func (c *listCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "list queue",
			Description: "list all queued commands",
			Examples: []string{
				"list queue",
				"list queue in channel",
			},
		},
	}
}
