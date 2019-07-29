package jira

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/command/queue"
	"github.com/innogames/slack-bot/config"
	"github.com/nlopes/slack"
	"gopkg.in/andygrunwald/go-jira.v1"
	"time"
)

// NewWatchCommand will inform the user abut the first ticket state change
func NewWatchCommand(jira *jira.Client, slackClient client.SlackClient, config config.Jira) bot.Command {
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
	ticketId := match.GetString("ticketId")
	issue, _, err := c.jira.Issue.Get(ticketId, nil)

	if err != nil {
		c.slackClient.Reply(event, err.Error())
		return
	}

	go c.watchTicket(event, issue)

	// add button to link
	c.slackClient.SendMessage(
		event,
		fmt.Sprintf("I'll inform you about changes of ticket %s", ticketId),
	)
}

func (c *watchCommand) watchTicket(event slack.MessageEvent, issue *jira.Issue) {
	lastStatus := issue.Fields.Status.Name
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	done := queue.AddRunningCommand(event, event.Text)
	for range ticker.C {
		issue, _, err := c.jira.Issue.Get(issue.ID, nil)
		if err != nil {
			done <- true
			c.slackClient.ReplyError(event, err)
			return
		}
		newStatus := issue.Fields.Status.Name

		if newStatus != lastStatus {
			c.slackClient.Reply(event, fmt.Sprintf(
				"%s %s: status changed from *%s* to *%s*",
				getFormattedUrl(c.config, issue),
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
			"watch jira",
			"inform you about changes jira states",
			[]string{
				"watch ticket PROJECT-1234",
			},
		},
	}
}
