package vcs

import (
	bitbucketServer "github.com/gfleury/go-bitbucket-v1/test/bb-mock-server/go"
	"github.com/innogames/slack-bot/bot/config"
	client2 "github.com/innogames/slack-bot/client"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBitbucketLoader(t *testing.T) {
	go func() {
		bitbucketServer.RunServer(7991)
	}()
	time.Sleep(time.Millisecond * 200)

	cfg := config.Bitbucket{
		Host:       "http://localhost:7991",
		Project:    "myProject",
		Repository: "myRepo",
		ApiKey:     "0815",
	}
	bitbucketClient, _ := client2.GetBitbucketClient(cfg)
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
