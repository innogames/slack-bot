package jira

import (
	"fmt"
	"github.com/andygrunwald/go-jira"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/command/queue"
	"github.com/slack-go/slack"
	"time"
)

// newWatchCommand will inform the user abut the first ticket state change
func newWatchCommand(jira *jira.Client, slackClient client.SlackClient, config config.Jira) bot.Command {
	return &watchCommand{jira, slackClient, config}
}

type watchCommand struct {
	jira        *jira.Client
	slackClient client.SlackClient
	config      config.Jira
}

func (c *watchCommand) IsEnabled() bool {
	return c.config.IsEnabled()
}

func (c *watchCommand) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher(`watch ticket (?P<ticketId>(\w+)-(\d+))`, c.Run)
}

func (c *watchCommand) Run(match matcher.Result, event slack.MessageEvent) {
	ticketID := match.GetString("ticketId")
	issue, response, err := c.jira.Issue.Get(ticketID, nil)

	if err != nil {
		c.slackClient.Reply(event, err.Error())
		return
	}
	response.Body.Close()

	go c.watchTicket(event, issue)

	// add button to link
	c.slackClient.SendMessage(
		event,
		fmt.Sprintf("I'll inform you about changes of ticket %s", ticketID),
	)
}

func (c *watchCommand) watchTicket(event slack.MessageEvent, issue *jira.Issue) {
	lastStatus := issue.Fields.Status.Name
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	done := queue.AddRunningCommand(event, event.Text)
	for range ticker.C {
		issue, resp, err := c.jira.Issue.Get(issue.ID, nil)
		if err != nil {
			done <- true
			c.slackClient.ReplyError(event, err)
			return
		}
		resp.Body.Close()
		newStatus := issue.Fields.Status.Name

		if newStatus != lastStatus {
			c.slackClient.Reply(event, fmt.Sprintf(
				"%s %s: status changed from *%s* to *%s*",
				getFormattedURL(c.config, *issue),
				issue.Fields.Summary,
				lastStatus,
				newStatus,
			))

			done <- true
			return
		}
	}
}

func (c *watchCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "watch jira",
			Description: "inform you about changes jira states",
			Examples: []string{
				"watch ticket PROJECT-1234",
			},
		},
	}
}
