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

func NewListCommand(slackClient client.SlackClient) bot.Command {
	return &listCommand{
		slackClient,
	}
}

func (c *listCommand) GetMatcher() matcher.Matcher {
	return matcher.NewTextMatcher("list queue", c.Run)
}

func (c *listCommand) Run(match matcher.Result, event slack.MessageEvent) {
	res, _ := storage.ReadAll(storageKey)
	response := fmt.Sprintf("%d queued commands\n", len(res))
	now := time.Now()

	var queuedEvent slack.MessageEvent
	for _, eventString := range res {
		if err := json.Unmarshal([]byte(eventString), &queuedEvent); err != nil {
			continue
		}

		userId, _ := client.GetUser(queuedEvent.User)
		i, _ := strconv.ParseInt(queuedEvent.Timestamp[0:10], 10, 64)
		t := time.Unix(i, 0)
		response += fmt.Sprintf(
			" - <@%s> (%s, %s ago): ```%s```  \n",
			userId,
			t.Format(time.Stamp),
			util.FormatDuration(now.Sub(t)),
			queuedEvent.Text,
		)
	}

	c.slackClient.Reply(event, response)
}

func (c *listCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "list queue",
			Description: "list all queued commands",
		},
	}
}
