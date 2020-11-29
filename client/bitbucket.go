package client

import (
	"context"
	"errors"
	bitbucket "github.com/gfleury/go-bitbucket-v1"
	"github.com/innogames/slack-bot/bot/config"
)

// GetBitbucketClient initialized a API client based on the given config
func GetBitbucketClient(cfg config.Bitbucket) (*bitbucket.DefaultApiService, error) {
	if !cfg.IsEnabled() {
		return nil, errors.New("bitbucket: No host given")
	}

	// todo add proper configurable timeout
	ctx := context.Background()

	if cfg.APIKey != "" {
		apiKey := bitbucket.APIKey{Key: cfg.APIKey}
		ctx = context.WithValue(ctx, bitbucket.APIKey{}, apiKey)
	} else if cfg.Username != "" && cfg.Password != "" {
		basicAuth := bitbucket.BasicAuth{UserName: cfg.Username, Password: cfg.Password}
		ctx = context.WithValue(ctx, bitbucket.ContextBasicAuth, basicAuth)
	}

	config := bitbucket.NewConfiguration(cfg.Host + "/rest")
	bitbucketClient := bitbucket.NewAPIClient(ctx, config)

	return bitbucketClient.DefaultApi, nil
}
