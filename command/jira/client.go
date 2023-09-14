package jira

import (
	"net/http"

	"github.com/andygrunwald/go-jira"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/client"
)

// getClient creates a jira client based on github.com/andygrunwald/go-jira and passes authentication credentials
func getClient(cfg *config.Jira) (*jira.Client, error) {
	var httpClient *http.Client

	switch {
	case cfg.Password != "":
		authClient := jira.BasicAuthTransport{
			Username: cfg.Username,
			Password: cfg.Password,
		}
		httpClient = authClient.Client()
	case cfg.AccessToken != "":
		authClient := jira.PATAuthTransport{
			Token: cfg.AccessToken,
		}
		httpClient = authClient.Client()
	default:
		// no authentication...
		httpClient = client.GetHTTPClient()
	}

	return jira.NewClient(httpClient, cfg.Host)
}
