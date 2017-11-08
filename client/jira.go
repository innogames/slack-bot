package client

import (
	"github.com/innogames/slack-bot/config"
	"gopkg.in/andygrunwald/go-jira.v1"
)

// GetJiraClient created a jira client based on gopkg.in/andygrunwald/go-jira.v1
func GetJiraClient(cfg config.Jira) (*jira.Client, error) {
	jiraTransport := jira.BasicAuthTransport{
		Username: cfg.Username,
		Password: cfg.Password,
	}

	return jira.NewClient(jiraTransport.Client(), cfg.Host)
}
