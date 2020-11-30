package jira

import (
	"fmt"
	"github.com/andygrunwald/go-jira"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/client"
	"github.com/slack-go/slack"
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

const (
	FormatDefault = "default"
	FormatFull    = "full"
	FormatLink    = "link"
)

var ticketRe = regexp.MustCompile(`^(\w+)-(\d+)$`)

// newJiraCommand search for a Jira ticket id or a JQL
func newJiraCommand(jira *jira.Client, slackClient client.SlackClient, config config.Jira) bot.Command {
	return &jiraCommand{jira, slackClient, config}
}

func (c *jiraCommand) IsEnabled() bool {
	return c.config.IsEnabled()
}

func (c *jiraCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewRegexpMatcher("jira (?P<action>link) (?P<text>.*)", c.Run),
		matcher.NewRegexpMatcher("(?P<action>jira|issue|jql) (?P<text>.*)", c.Run),
	)
}

func (c *jiraCommand) Run(match matcher.Result, message msg.Message) {
	eventText := match.GetString("text")
	ticketNumber := c.getTicketNumber(eventText)
	action := match.GetString("action")

	if ticketNumber != "" {
		issue, _, err := c.jira.Issue.Get(ticketNumber, nil)
		if err != nil {
			c.slackClient.SendMessage(message, err.Error())
			return
		}

		c.sendTicket(message, issue, action)
		return
	}

	// search by JQL
	defaultProject := c.config.Project
	var jql string
	if action == "jql" {
		jql = match.GetString("text")
		if !strings.Contains(jql, "project") {
			// search in default project
			jql = fmt.Sprintf("project = %s AND %s", defaultProject, jql)
		}
	} else {
		// search in default project
		jql = fmt.Sprintf("project = %s AND text ~ '%s' ORDER BY priority DESC", defaultProject, eventText)
	}

	c.jqlList(message, jql)
}

// "1234" -> PROJ-1234
// "FOO-1234" -> "FOO-1234"
// dsadsad -> ""
func (c *jiraCommand) getTicketNumber(eventText string) string {
	var ticketNumber string
	if _, err := strconv.Atoi(eventText); err == nil {
		ticketNumber = c.config.Project + "-" + eventText
	} else if ticketRe.MatchString(eventText) {
		ticketNumber = eventText
	}
	return ticketNumber
}

func (c *jiraCommand) sendTicket(ref msg.Ref, issue *jira.Issue, format string) {
	if format == FormatLink {
		text := fmt.Sprintf("<%s|%s: %s>", getTicketURL(c.config, *issue), issue.Key, issue.Fields.Summary)
		c.slackClient.SendMessage(ref, text)
		return
	}

	fields := []slack.AttachmentField{
		{
			Title: "Name",
			Value: fmt.Sprintf("%s: %s", getFormattedURL(c.config, *issue), issue.Fields.Summary),
		},
		{
			Title: "Priority",
			Value: c.getField("Priority", issue.Fields.Priority.Name),
			Short: true,
		},
		{
			Title: "Type",
			Value: c.getField("Type", issue.Fields.Type.Name),
			Short: true,
		},
	}

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

	if len(issue.Fields.Labels) > 0 {
		fields = append(fields, slack.AttachmentField{
			Title: "Labels",
			Value: strings.Join(issue.Fields.Labels, ", "),
			Short: false,
		})
	}

	attachment := slack.Attachment{
		Color:  "#D3D3D3",
		Fields: fields,
		MarkdownIn: []string{
			"text", "fields",
		},
		Actions: []slack.AttachmentAction{
			client.GetSlackLink("Open in Jira", getTicketURL(c.config, *issue)),
		},
	}

	c.slackClient.SendMessage(ref, "", slack.MsgOptionAttachments(attachment))
}

// get the name of the field and the mapped icon (if configured)
func (c *jiraCommand) getField(fieldType string, name string) string {
	for _, field := range c.config.Fields {
		if field.Name == fieldType {
			if icon, ok := field.Icons[name]; ok {
				return fmt.Sprintf("%s %s", name, icon)
			}
			return name + " :question:" // todo const
		}
	}

	return name
}

func (c *jiraCommand) jqlList(message msg.Message, jql string) {
	search := &url.URL{Path: jql}

	tickets, _, err := c.jira.Issue.Search(jql, nil)
	if err != nil {
		c.slackClient.SendMessage(message, err.Error())
		return
	}

	if len(tickets) == 1 {
		c.sendTicket(message, &tickets[0], FormatDefault)
		return
	}

	text := fmt.Sprintf("I found %d matching ticket(s).\n", len(tickets))
	for _, ticket := range tickets {
		if ticket.Fields == nil {
			continue
		}
		text += fmt.Sprintf(
			"%s %s%s - %s (%s)",
			getFormattedURL(c.config, ticket),
			idToIcon(ticket.Fields.Priority),
			c.getField("Type", ticket.Fields.Type.Name),
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

	c.slackClient.SendMessage(message, text, slack.MsgOptionAttachments(attachment))
}

func (c *jiraCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "jira",
			Description: "list jira ticket information or performs jira searches. It uses the configured jira project by default to display/search tickets",
			Category:    category,
			Examples: []string{
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
