package client

import (
	"github.com/andygrunwald/go-jira"
	"github.com/innogames/slack-bot/bot/config"
)

// GetJiraClient created a jira client based on github.com/andygrunwald/go-jira"
func GetJiraClient(cfg *config.Jira) (*jira.Client, error) {
	client := GetHTTPClient()

	if cfg.Username != "" {
		authClient := jira.BasicAuthTransport{
			Username: cfg.Username,
			Password: cfg.Password,
		}
		client = authClient.Client()
	}

	return jira.NewClient(client, cfg.Host)
}
