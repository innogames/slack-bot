package vcs

import (
	"testing"
	"time"

	bitbucketServer "github.com/gfleury/go-bitbucket-v1/test/bb-mock-server/go"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/client"
	"github.com/stretchr/testify/assert"
)

func TestBitbucketLoader(t *testing.T) {
	go bitbucketServer.RunServer(7991)

	// todo(matze) defer shutdown
	// todo(matze) wait till server active

	time.Sleep(time.Millisecond * 100)

	cfg := config.Bitbucket{
		Host:       "http://localhost:7991",
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
		assert.Equal(t, []string{"feature/branch"}, branches)
	})
}
