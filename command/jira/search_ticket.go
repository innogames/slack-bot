package jira

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/config"
	"github.com/nlopes/slack"
	"gopkg.in/andygrunwald/go-jira.v1"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

type jiraCommand struct {
	jira        *jira.Client
	slackClient client.SlackClient
	config      config.Jira
}

var ticketRe = regexp.MustCompile(`^(\w+)-(\d+)$`)

// NewJiraCommand search for a Jira ticket id or a JQL
func NewJiraCommand(jira *jira.Client, slackClient client.SlackClient, config config.Jira) bot.Command {
	return &jiraCommand{jira, slackClient, config}
}

func (c *jiraCommand) IsEnabled() bool {
	return c.config.IsEnabled()
}

func (c *jiraCommand) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher("(?P<action>jira|issue|jql) (?P<text>.*)", c.Run)
}

func (c *jiraCommand) Run(match matcher.Result, event slack.MessageEvent) {
	eventText := match.GetString("text")
	ticketNumber := c.getTicketNumber(eventText)

	if ticketNumber != "" {
		issue, response, err := c.jira.Issue.Get(ticketNumber, nil)

		if response == nil || response.StatusCode > 400 {
			c.slackClient.Reply(event, err.Error())
			return
		}
		c.sendTicket(event, issue)
		return
	}

	// search by JQL
	defaultProject := c.config.Project
	var jql string
	if match.GetString("action") == "jql" {
		jql = match.GetString("text")
		if !strings.Contains(jql, "project") {
			// search in default project
			jql = fmt.Sprintf("project = %s AND %s", defaultProject, jql)
		}
	} else {
		// search in default project
		jql = fmt.Sprintf("project = %s AND text ~ '%s' ORDER BY priority DESC", defaultProject, eventText)
	}

	c.jqlList(event, jql)
}

// "1234" -> PROJ-1234
func (c *jiraCommand) getTicketNumber(eventText string) string {
	var ticketNumber string
	if _, err := strconv.Atoi(eventText); err == nil {
		ticketNumber = c.config.Project + "-" + eventText
	} else if ticketRe.MatchString(eventText) {
		ticketNumber = eventText
	}
	return ticketNumber
}

func (c *jiraCommand) sendTicket(event slack.MessageEvent, issue *jira.Issue) {
	information := idToIcon(issue.Fields.Priority)
	information += " " + issue.Fields.Type.Name + typeIcon(issue.Fields.Type.Name)

	var fields []slack.AttachmentField
	fields = append(
		fields,
		slack.AttachmentField{
			Title: "Name",
			Value: fmt.Sprintf("%s: %s", getTicketUrl(c.config, issue), issue.Fields.Summary),
		},
		slack.AttachmentField{
			Title: "Information",
			Value: information,
		},
	)

	if issue.Fields.Assignee != nil {
		fields = append(fields, slack.AttachmentField{
			Title: "Assignee",
			Value: fmt.Sprintf(
				"<%sissues/?jql=assignee=\"%s\" AND status != \"Closed\"|%s>",
				c.config.Host,
				issue.Fields.Assignee.Name,
				issue.Fields.Assignee.DisplayName,
			),
			Short: true,
		})
	}

	fields = append(fields, slack.AttachmentField{
		Title: "Status",
		Value: issue.Fields.Status.Name,
		Short: true,
	})

	if len(issue.Fields.Components) > 0 {
		var components []string
		for _, component := range issue.Fields.Components {
			components = append(components, component.Name)
		}
		fields = append(fields, slack.AttachmentField{
			Title: "Components",
			Value: strings.Join(components, ", "),
			Short: true,
		})
	}

	fields = append(fields, slack.AttachmentField{
		Title: "Description",
		Value: convertMarkdown(issue.Fields.Description),
		Short: false,
	})

	attachment := slack.Attachment{
		Color:  "#D3D3D3",
		Fields: fields,
		MarkdownIn: []string{
			"text", "fields",
		},
	}

	c.slackClient.SendMessage(event, "", slack.MsgOptionAttachments(attachment))
}

func (c *jiraCommand) jqlList(event slack.MessageEvent, jql string) {
	search := &url.URL{Path: jql}

	tickets, _, err := c.jira.Issue.Search(jql, nil)
	if err != nil {
		c.slackClient.Reply(event, err.Error())
		return
	}

	if len(tickets) == 1 {
		c.sendTicket(event, &tickets[0])
		return
	}

	text := fmt.Sprintf("I found %d matching ticket(s).\n", len(tickets))
	for _, ticket := range tickets {
		if ticket.Fields == nil {
			continue
		}
		text += fmt.Sprintf(
			"%s %s%s - %s (%s)",
			getTicketUrl(c.config, &ticket),
			idToIcon(ticket.Fields.Priority),
			typeIcon(ticket.Fields.Type.Name),
			ticket.Fields.Summary,
			ticket.Fields.Status.Name,
		) + "\n"
	}

	// add button which leads to search
	searchLink := fmt.Sprintf("%sissues/?jql=%s", c.config.Host, search.String())
	attachment := slack.Attachment{}
	attachment.Actions = append(
		attachment.Actions,
		client.GetSlackLink("Search in Jira", searchLink),
	)

	c.slackClient.SendMessage(event, text, slack.MsgOptionAttachments(attachment))
}

func (c *jiraCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			"jira",
			"list jira ticket information or performs jira searches. It uses the configured jira project by default to display/search tickets",
			[]string{
				"jql",
				"jql status=\"In Progress\"",
				"issue 43234",
				"issue PROJ-23123",
				"issue \"second city\"",
			},
		},
	}
}

func (c *jiraCommand) GetTemplateFunction() template.FuncMap {
	return template.FuncMap{
		"jiraTicket": func(ticketId string) *jira.Issue {
			issue, _, _ := c.jira.Issue.Get(ticketId, nil)

			return issue
		},
		"jiraTicketUrl": func(ticketId string) string {
			return fmt.Sprintf("%sbrowse/%s", c.config.Host, ticketId)
		},
	}
}
