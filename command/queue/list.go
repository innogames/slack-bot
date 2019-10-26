package queue

import (
	"encoding/json"
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/bot/storage"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/innogames/slack-bot/client"
	"github.com/nlopes/slack"
	"strconv"
	"time"
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
	response := ""

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
		i, _ := strconv.ParseInt(queuedEvent.Timestamp[0:10], 10, 64)
		t := time.Unix(i, 0)
		response += fmt.Sprintf(
			" - <@%s> (%s, %s ago): ```%s``` %s \n",
			userId,
			t.Format(time.Stamp),
			util.FormatDuration(now.Sub(t)),
			queuedEvent.Text,
			c.getReactions(queuedEvent),
		)
	}

	response = fmt.Sprintf("%d queued commands\n", count) + response

	c.slackClient.Reply(event, response)
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
