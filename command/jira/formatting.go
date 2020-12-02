package jira

import (
	"fmt"
	"github.com/andygrunwald/go-jira"
	"github.com/innogames/slack-bot/bot/config"
	"strings"
)

// todo map all icons in config (Priority.Name.Blocker = :jira_blocker: ...
// todo move to default map
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

func convertMarkdown(content string) string {
	content = strings.ReplaceAll(content, "{code}", "```")
	content = strings.ReplaceAll(content, "{quote}", "```")
	content = strings.ReplaceAll(content, "h1. ", "")
	content = strings.ReplaceAll(content, "h2. ", "")
	content = strings.ReplaceAll(content, "h3. ", "")

	return content
}

func getFormattedURL(cfg *config.Jira, issue jira.Issue) string {
	return fmt.Sprintf("<%s|%s>", getTicketURL(cfg, issue), issue.Key)
}

func getTicketURL(cfg *config.Jira, issue jira.Issue) string {
	return fmt.Sprintf("%sbrowse/%s", cfg.Host, issue.Key)
}
