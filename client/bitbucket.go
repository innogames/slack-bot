package client

import (
	"context"
	"errors"
	bitbucket "github.com/gfleury/go-bitbucket-v1"
	"github.com/innogames/slack-bot/bot/config"
)

func GetBitbucketClient(cfg config.Bitbucket) (*bitbucket.APIClient, error) {
	if !cfg.IsEnabled() {
		return nil, errors.New("bitbucket: No host given")
	}

	// todo add proper configurable timeout
	ctx := context.Background()
	if cfg.ApiKey != "" {
		apiKey := bitbucket.APIKey{Key: cfg.ApiKey}
		ctx = context.WithValue(ctx, bitbucket.APIKey{}, apiKey)
	} else if cfg.Username != "" && cfg.Password != "" {
		basicAuth := bitbucket.BasicAuth{UserName: cfg.Username, Password: cfg.Password}
		ctx = context.WithValue(ctx, bitbucket.ContextBasicAuth, basicAuth)
	} else {
		return nil, errors.New("bitbucket: No username/password or api_key given")
	}

	config := bitbucket.NewConfiguration(cfg.Host + "/rest")
	bitbucketClient := bitbucket.NewAPIClient(ctx, config)

	return bitbucketClient, nil
}