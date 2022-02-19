package client

import (
	"net/http"

	"github.com/andygrunwald/go-jira"
	"github.com/innogames/slack-bot/v2/bot/config"
)

// GetJiraClient creates a jira client based on github.com/andygrunwald/go-jira and passes authentication credentials
func GetJiraClient(cfg *config.Jira) (*jira.Client, error) {
	var client *http.Client

	switch {
	case cfg.Password != "":
		authClient := jira.BasicAuthTransport{
			Username: cfg.Username,
			Password: cfg.Password,
		}
		client = authClient.Client()
	case cfg.AccessToken != "":
		authClient := jira.PATAuthTransport{
			Token: cfg.AccessToken,
		}
		client = authClient.Client()
	default:
		// no authentication...
		client = GetHTTPClient()
	}

	return jira.NewClient(client, cfg.Host)
}
