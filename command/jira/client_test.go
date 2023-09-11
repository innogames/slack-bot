package jira

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/stretchr/testify/assert"
)

func TestJiraClient(t *testing.T) {
	t.Run("no credentials", func(t *testing.T) {
		cfg := &config.Jira{
			Host: "https://jira.example.com",
		}
		client, err := getClient(cfg)

		assert.Nil(t, err)
		assert.Equal(t, "jira.example.com", client.GetBaseURL().Host)
	})

	t.Run("with password", func(t *testing.T) {
		cfg := &config.Jira{
			Host:     "https://jira.example.com",
			Username: "foo",
			Password: "bar",
		}
		client, err := getClient(cfg)

		assert.Nil(t, err)
		assert.False(t, client.Authentication.Authenticated())
	})

	t.Run("with access token", func(t *testing.T) {
		cfg := &config.Jira{
			Host:        "https://jira.example.com",
			Username:    "foo",
			AccessToken: "iamsecret",
		}
		client, err := getClient(cfg)

		assert.Nil(t, err)
		assert.False(t, client.Authentication.Authenticated())
		client.Authentication.GetCurrentUser()
	})
}
