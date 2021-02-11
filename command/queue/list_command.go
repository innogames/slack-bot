package queue

import (
	"fmt"
	"time"

	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/storage"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/innogames/slack-bot/client"
	"github.com/slack-go/slack"
)

type listCommand struct {
	bot.BaseCommand
}

type filterFunc func(msg.Message) bool

// NewListCommand prints the list of all queued commands (blocking commands like running Jenkins jobs)
func NewListCommand(base bot.BaseCommand) bot.Command {
	return &listCommand{
		base,
	}
}

func (c *listCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewTextMatcher("list queue", c.ListAll),
		matcher.NewTextMatcher("list queue in channel", c.ListChannel),
	)
}

func (c *listCommand) ListAll(match matcher.Result, message msg.Message) {
	msgOptions := c.listQueue(func(event msg.Message) bool {
		return true
	})
	c.SendMessage(message, "", msgOptions)
}

func (c *listCommand) ListChannel(match matcher.Result, message msg.Message) {
	msgOptions := c.listQueue(func(queuedEvent msg.Message) bool {
		return message.GetChannel() == queuedEvent.GetChannel()
	})
	c.SendMessage(message, "", msgOptions)
}

func (c *listCommand) listQueue(filter filterFunc) slack.MsgOption {
	now := time.Now()
	count := 0

	keys, _ := storage.GetKeys(storageKey)
	attachments := make([]slack.Attachment, 0, len(keys)+1)

	var queuedEvent msg.Message
	for _, key := range keys {
		if err := storage.Read(storageKey, key, &queuedEvent); err != nil {
			continue
		}

		if !filter(queuedEvent) {
			continue
		}

		count++
		_, userName := client.GetUser(queuedEvent.User)

		messageTime := queuedEvent.GetTime()
		timeAgo := now.Sub(messageTime)
		color := getColor(timeAgo)
		text := fmt.Sprintf(
			"*%s* (<%s|%s, %s ago>): ```%s``` %s\n",
			userName,
			client.GetSlackArchiveLink(queuedEvent),
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

	// prepend the number of matched commands (after filtering :))
	attachments = append([]slack.Attachment{
		{
			Text: fmt.Sprintf("*%d queued commands*", count),
		},
	}, attachments...)

	return slack.MsgOptionAttachments(attachments...)
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
	reactions, _ := c.GetReactions(msgRef, slack.NewGetReactionsParameters())

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
