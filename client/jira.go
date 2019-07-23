package client

import (
	"github.com/innogames/slack-bot/config"
	"gopkg.in/andygrunwald/go-jira.v1"
	"net/http"
)

// GetJiraClient created a jira client based on gopkg.in/andygrunwald/go-jira.v1
func GetJiraClient(cfg config.Jira) (*jira.Client, error) {
	var client *http.Client

	if cfg.Username != "" {
		authClient := jira.BasicAuthTransport{
			Username: cfg.Username,
			Password: cfg.Password,
		}
		client = authClient.Client()
	}

	return jira.NewClient(client, cfg.Host)
}
