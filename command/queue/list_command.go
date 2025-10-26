package queue

import (
	"fmt"
	"text/template"
	"time"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/storage"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/slack-go/slack"
)

const (
	processingReaction = "eyes"
	defaultName        = "queued commands"
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
		matcher.NewTextMatcher("list queue", c.listAll),
		matcher.NewOptionMatcher("list queue in channel", []string{"pin", "hide-empty", "element-name"}, c.listChannel, c.SlackClient),
	)
}

// ListAll shows a list of all queued commands
func (c *listCommand) listAll(match matcher.Result, message msg.Message) {
	c.sendList(message, match, func(_ msg.Message) bool {
		return true
	})
}

// ListChannel shows a list of all queued commands within the current channel
func (c *listCommand) listChannel(match matcher.Result, message msg.Message) {
	c.sendList(message, match, func(queuedEvent msg.Message) bool {
		return message.GetChannel() == queuedEvent.GetChannel()
	})
}

// format a block-based message with all matching commands
func (c *listCommand) sendList(message msg.Message, options matcher.Result, filter filterFunc) {
	// add :eyes: temporary because loading the list might take some seconds
	c.AddReaction(processingReaction, message)
	defer c.RemoveReaction(processingReaction, message)

	count, queueBlocks := c.getQueueAsBlocks(message, filter)

	if count == 0 && options.Has("hide-empty") {
		// hide empty list
		return
	}

	elementName := defaultName
	if options.Has("element-name") {
		elementName = options.GetString("element-name")
	}

	blocks := []slack.Block{
		client.GetTextBlock(fmt.Sprintf("*%d %s*", count, elementName)),
	}
	blocks = append(blocks, queueBlocks...)

	// replace the original message when it's triggered by the "refresh" button
	var msgOptions []slack.MsgOption
	if message.IsUpdatedMessage() {
		msgOptions = append(msgOptions, slack.MsgOptionUpdate(message.Timestamp))
	}

	messageTimestamp := c.SendBlockMessage(message, blocks, msgOptions...)

	if options.Has("pin") {
		err := c.PinMessage(message.Channel, messageTimestamp)
		c.ReplyError(message, err)
	}
}

// loads all matching queue entries and format them into slack.Block
func (c *listCommand) getQueueAsBlocks(message msg.Message, filter filterFunc) (count uint, blocks []slack.Block) {
	now := time.Now()
	keys, _ := storage.GetKeys(storageKey)

	var queuedEvent msg.Message
	for _, key := range keys {
		if err := storage.Read(storageKey, key, &queuedEvent); err != nil {
			continue
		}

		if !filter(queuedEvent) {
			continue
		}

		count++
		_, userName := client.GetUserIDAndName(queuedEvent.User)

		messageTime := queuedEvent.GetTime()
		timeAgo := now.Sub(messageTime)
		text := fmt.Sprintf(
			"*%s* (<%s|%s, %s ago>): ```%s``` %s",
			userName,
			client.GetSlackArchiveLink(queuedEvent),
			messageTime.Format(time.Stamp),
			util.FormatDuration(timeAgo),
			queuedEvent.GetText(),
			c.getReactions(queuedEvent),
		)

		textBlock := client.GetTextBlock(text)
		blocks = append(
			blocks,
			textBlock,
		)
	}

	// add "Updated at..." time if there was an update
	if message.IsUpdatedMessage() {
		blocks = append(
			blocks,
			client.GetContextBlock("Updated at: "+time.Now().Format(time.Stamp)),
		)
	}

	// add "Refresh" button
	blocks = append(
		blocks,
		slack.NewActionBlock(
			"",
			client.GetInteractionButton("refresh", "Refresh :arrows_counterclockwise:", message.GetText()),
		),
	)

	return count, blocks
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

func (c *listCommand) GetTemplateFunction() template.FuncMap {
	return template.FuncMap{
		"countBackgroundJobs": func() int {
			keys, _ := storage.GetKeys(storageKey)

			return len(keys)
		},
		"countBackgroundJobsInChannel": func(channel string) int {
			count := 0
			keys, _ := storage.GetKeys(storageKey)
			var queuedEvent msg.Message
			for _, key := range keys {
				if err := storage.Read(storageKey, key, &queuedEvent); err != nil {
					continue
				}
				if queuedEvent.Channel != channel {
					continue
				}
				count++
			}

			return count
		},
	}
}

func (c *listCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "list queue",
			Description: "list all queued commands",
			Examples: []string{
				"list queue",
			},
		},
		{
			Command:     "list queue in channel",
			Description: "list queued commands in current channel",
			Examples: []string{
				"list queue in channel",
				"list queue in channel pin=true",
			},
		},
	}
}
