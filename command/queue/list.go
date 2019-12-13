package queue

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/bot/storage"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/innogames/slack-bot/client"
	"github.com/nlopes/slack"
)

type listCommand struct {
	slackClient client.SlackClient
}

type filterFunc func(slack.MessageEvent) bool

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

func (c *listCommand) ListAll(match matcher.Result, event slack.MessageEvent) {
	c.listQueue(match, event, func(event slack.MessageEvent) bool {
		return true
	})
}

func (c *listCommand) ListChannel(match matcher.Result, event slack.MessageEvent) {
	c.listQueue(match, event, func(queuedEvent slack.MessageEvent) bool {
		return event.Channel == queuedEvent.Channel
	})
}

func (c *listCommand) listQueue(match matcher.Result, event slack.MessageEvent, filter filterFunc) {
	res, _ := storage.ReadAll(storageKey)
	now := time.Now()

	count := 0
	attachments := make([]slack.Attachment, 0, len(res))

	var queuedEvent slack.MessageEvent
	for _, eventString := range res {
		if err := json.Unmarshal([]byte(eventString), &queuedEvent); err != nil {
			continue
		}

		if !filter(queuedEvent) {
			continue
		}

		count++
		userId, _ := client.GetUser(queuedEvent.User)

		messageTime := util.GetMessageTime(queuedEvent)
		timeAgo := now.Sub(messageTime)
		color := getColor(timeAgo)
		text := fmt.Sprintf(
			"<@%s> (%s, %s ago): ```%s``` %s \n",
			userId,
			messageTime.Format(time.Stamp),
			util.FormatDuration(timeAgo),
			queuedEvent.Text,
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

	c.slackClient.SendMessage(event, response, slack.MsgOptionAttachments(attachments...))
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

func (c *listCommand) getReactions(event slack.MessageEvent) string {
	formattedReactions := ""
	msgRef := slack.NewRefToMessage(event.Channel, event.Timestamp)
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
