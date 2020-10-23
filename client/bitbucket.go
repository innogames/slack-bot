package client

import (
	"context"
	bitbucket "github.com/gfleury/go-bitbucket-v1"
	"github.com/innogames/slack-bot/bot/config"
	"time"
)

func GetBitbucketClient(cfg config.Bitbucket) (*bitbucket.APIClient, error) {
	basicAuth := bitbucket.BasicAuth{UserName: cfg.Username, Password: cfg.Password}
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	ctx = context.WithValue(ctx, bitbucket.ContextBasicAuth, basicAuth)

	config := bitbucket.NewConfiguration(cfg.Host + "/rest")
	bitbucketClient := bitbucket.NewAPIClient(ctx, config)

	return bitbucketClient, nil
}
