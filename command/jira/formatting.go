package jira

import (
	"fmt"
	"github.com/innogames/slack-bot/config"
	"gopkg.in/andygrunwald/go-jira.v1"
	"strings"
)

func idToIcon(priority *jira.Priority) string {
	if priority == nil {
		return ":question:"
	}

	switch priority.Name {
	case "Blocker":
		return ":jira_blocker:"
	case "Critical":
		return ":jira_critical:"
	case "Major":
		return ":jira_major:"
	case "Medium":
		return ":jira_medium:"
	case "Minor":
		return ":jira_minor:"
	default:
		return ":white_circle:"
	}
}

func typeIcon(ticketType string) string {
	switch ticketType {
	case "Bug":
		return ":beetle:"
	}
	return ""
}

func convertMarkdown(content string) string {
	content = strings.Replace(content, "{code}", "```", -1)
	content = strings.Replace(content, "{quote}", "```", -1)
	content = strings.Replace(content, "h1. ", "", -1)
	content = strings.Replace(content, "h2. ", "", -1)
	content = strings.Replace(content, "h3. ", "", -1)

	return content
}

func getTicketUrl(cfg config.Jira, issue *jira.Issue) string {
	return fmt.Sprintf("<%sbrowse/%s|%s>", cfg.Host, issue.Key, issue.Key)
}
