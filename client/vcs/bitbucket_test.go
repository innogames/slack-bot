package vcs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/stretchr/testify/assert"
)

func TestBitbucketLoader(t *testing.T) {
	server := spawnBitbucketTestServer()
	defer server.Close()

	t.Run("Load branches", func(t *testing.T) {
		cfg := config.Bitbucket{
			Host:       server.URL,
			Project:    "myProject",
			Repository: "myRepo",
			APIKey:     "0815",
		}
		bitbucketClient, _ := client.GetBitbucketClient(cfg)
		fetcher := &bitbucket{
			bitbucketClient,
			cfg,
		}

		branches, err := fetcher.LoadBranches()
		assert.Nil(t, err)
		assert.Equal(t, []string{"master", "release"}, branches)
	})

	t.Run("Load branches with not existing repo", func(t *testing.T) {
		cfg := config.Bitbucket{
			Host:       server.URL,
			Project:    "myNotExisting",
			Repository: "myRepo",
			APIKey:     "0815",
		}
		bitbucketClient, _ := client.GetBitbucketClient(cfg)
		fetcher := &bitbucket{
			bitbucketClient,
			cfg,
		}

		branches, err := fetcher.LoadBranches()
		assert.True(t, strings.HasPrefix(err.Error(), "Status: 404 Not Found, Body: 404 page not found"))
		assert.Len(t, branches, 0)
	})
}

func spawnBitbucketTestServer() *httptest.Server {
	mux := http.NewServeMux()

	// 1337: merged pr
	mux.HandleFunc("/rest/api/1.0/projects/myProject/repos/myRepo/branches", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte(`{
			"values": [
				{
					"displayId": "master"
				},
				{
					"displayId": "release"
				}
			]
		}`))
	})

	return httptest.NewServer(mux)
}
