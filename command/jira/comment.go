package jira

import (
	"fmt"

	"github.com/andygrunwald/go-jira"
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/client"
)

// newCommentCommand adds a comment to the given ticket
func newCommentCommand(jiraClient *jira.Client, slackClient client.SlackClient, cfg *config.Jira) bot.Command {
	return &commentCommand{jiraClient, slackClient, cfg}
}

type commentCommand struct {
	jira        *jira.Client
	slackClient client.SlackClient
	config      *config.Jira
}

func (c *commentCommand) IsEnabled() bool {
	return c.config.IsEnabled()
}

func (c *commentCommand) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher(`add comment to ticket (?P<ticketId>(\w+)-(\d+)) (?P<comment>.+)`, c.addComment)
}

func (c *commentCommand) addComment(match matcher.Result, message msg.Message) {
	ticketID := match.GetString("ticketId")
	issue, _, err := c.jira.Issue.Get(ticketID, nil)
	if err != nil {
		c.slackClient.ReplyError(message, fmt.Errorf("invalid ticket: %s", ticketID))
		return
	}

	_, userName := client.GetUserIDAndName(message.GetUser())

	comment := fmt.Sprintf(
		"%s: %s",
		userName,
		match.GetString("comment"),
	)
	_, _, err = c.jira.Issue.AddComment(issue.Key, &jira.Comment{
		Body: comment,
	})
	if err != nil {
		c.slackClient.ReplyError(message, err)
		return
	}

	c.slackClient.AddReaction("✅", message)
}

func (c *commentCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "comment jira ticket <ticket> <comment>",
			Description: "adds a comment to a jira ticket",
			Category:    category,
			Examples: []string{
				"comment ticket PROJECT-1234 Please check it on test server, I fixed it!",
			},
		},
	}
}
