package jira

import (
	"fmt"

	"github.com/andygrunwald/go-jira"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/slack-go/slack"
)

// newCommentCommand adds a comment to the given ticket
func newCommentCommand(jira *jira.Client, slackClient client.SlackClient, config config.Jira) bot.Command {
	return &commentCommand{jira, slackClient, config}
}

type commentCommand struct {
	jira        *jira.Client
	slackClient client.SlackClient
	config      config.Jira
}

func (c *commentCommand) IsEnabled() bool {
	return c.config.IsEnabled()
}

func (c *commentCommand) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher(`add comment to ticket (?P<ticketId>(\w+)-(\d+)) (?P<comment>.+)`, c.AddComment)
}

func (c *commentCommand) AddComment(match matcher.Result, event slack.MessageEvent) {
	ticketID := match.GetString("ticketID")
	issue, resp, err := c.jira.Issue.Get(ticketID, nil)

	if err != nil {
		c.slackClient.ReplyError(event, err)
		return
	}
	resp.Body.Close()

	_, userName := client.GetUser(event.User)

	comment := fmt.Sprintf(
		"%s: %s",
		userName,
		match.GetString("comment"),
	)
	_, resp, err = c.jira.Issue.AddComment(issue.ID, &jira.Comment{
		Body: comment,
	})
	if err != nil {
		c.slackClient.ReplyError(event, err)
		return
	}
	resp.Body.Close()

	msgRef := slack.NewRefToMessage(event.Channel, event.Timestamp)
	c.slackClient.AddReaction("white_check_mark", msgRef)
}

func (c *commentCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "comment jira ticket",
			Description: "adds a comment to a jira ticket",
			Examples: []string{
				"comment ticket PROJECT-1234 Please check it on test server, I fixed it!",
			},
		},
	}
}
