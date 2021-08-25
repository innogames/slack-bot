package jira

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/andygrunwald/go-jira"
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/slack-go/slack"
)

type jiraCommand struct {
	jira        *jira.Client
	slackClient client.SlackClient
	config      *config.Jira
}

const (
	formatLink = "link"
)

var ticketRe = regexp.MustCompile(`^(\w+)-(\d+)$`)

// newJiraCommand search for a Jira ticket id or a JQL
func newJiraCommand(jiraClient *jira.Client, slackClient client.SlackClient, cfg *config.Jira) bot.Command {
	return &jiraCommand{jiraClient, slackClient, cfg}
}

func (c *jiraCommand) IsEnabled() bool {
	return c.config.IsEnabled()
}

func (c *jiraCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewRegexpMatcher("jira (?P<action>link) (?P<text>.*)", c.run),
		matcher.NewRegexpMatcher("(?P<action>jira|issue|jql) (?P<text>.*)", c.run),
		matcher.NewRegexpMatcher(
			fmt.Sprintf(
				"%s\\/browse\\/(?P<text>.*)",
				regexp.QuoteMeta(strings.TrimRight(c.config.Host, "/")),
			),
			c.run,
		),
	)
}

func (c *jiraCommand) run(match matcher.Result, message msg.Message) {
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
	if format == formatLink {
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

	listText := ""
	for _, ticket := range tickets {
		if ticket.Fields == nil {
			continue
		}
		listText += fmt.Sprintf(
			"%s %s%s - %s (%s - %s)\n",
			getFormattedURL(c.config, ticket),
			idToIcon(ticket.Fields.Priority),
			c.getField("Type", ticket.Fields.Type.Name),
			ticket.Fields.Summary,
			ticket.Fields.Status.Name,
			getAssignee(ticket.Fields.Assignee),
		)
	}

	// add button which leads to search
	searchLink := fmt.Sprintf("%sissues/?jql=%s", c.config.Host, search.String())
	attachment := slack.Attachment{}
	attachment.Title = fmt.Sprintf("I found <%s|%d matching ticket(s)>.\n", searchLink, len(tickets))
	attachment.Text = listText
	attachment.Actions = []slack.AttachmentAction{
		client.GetSlackLink("Search in Jira", searchLink),
	}

	c.slackClient.SendMessage(message, "", slack.MsgOptionAttachments(attachment))
}

func getAssignee(user *jira.User) string {
	if user == nil {
		return "unassigned"
	}

	return user.Name
}

func (c *jiraCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "jira <ticket|search>",
			Description: "list jira ticket information or performs jira searches. It uses the configured jira project by default to display/search tickets",
			Category:    category,
			Examples: []string{
				"issue 43234",
				fmt.Sprintf("%s/browse/PROJ-1234", strings.TrimRight(c.config.Host, "/")),
				"issue PROJ-23123",
				"issue \"second city\"",
			},
		},
		{
			Command:     "jql <query>",
			Description: "list jira tickets based on the given JQL",
			Category:    category,
			Examples: []string{
				"jql status=\"In Progress\"",
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
