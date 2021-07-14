package vcs

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/innogames/slack-bot.v2/bot/config"
	"github.com/innogames/slack-bot.v2/client"
	"github.com/stretchr/testify/assert"
)

func TestBitbucketLoader(t *testing.T) {
	server := spawnBitbucketTestServer()
	defer server.Close()

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

	t.Run("Load branches", func(t *testing.T) {
		branches, err := fetcher.LoadBranches()
		assert.Nil(t, err)
		assert.Equal(t, []string{"master", "release"}, branches)
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
