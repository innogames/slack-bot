package jira

import (
	"fmt"
	"strings"

	"github.com/andygrunwald/go-jira"
	"github.com/innogames/slack-bot/v2/bot/config"
)

// replace Jira markdown to valid slack message
var markdownReplacer = strings.NewReplacer(
	"{code}", "```",
	"{quote}", "```",
	"h1. ", "",
	"h2. ", "",
	"h3. ", "",
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
	return markdownReplacer.Replace(content)
}

func getFormattedURL(cfg *config.Jira, issue jira.Issue) string {
	return fmt.Sprintf("<%s|%s>", getTicketURL(cfg, issue), issue.Key)
}

func getTicketURL(cfg *config.Jira, issue jira.Issue) string {
	return fmt.Sprintf("%sbrowse/%s", cfg.Host, issue.Key)
}
