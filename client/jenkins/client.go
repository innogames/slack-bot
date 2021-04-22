package jenkins

//go:generate $GOPATH/bin/mockery --output ../../mocks --name Client

import (
	"context"
	"github.com/bndr/gojenkins"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/client"
)

// Client is a interface of gojenkins.Jenkins
type Client interface {
	GetJob(ctx context.Context, id string, parentIDs ...string) (*gojenkins.Job, error)
	BuildJob(ctx context.Context, name string, params map[string]string) (int64, error)
	GetAllNodes(ctx context.Context) ([]*gojenkins.Node, error)
}

// GetClient created Jenkins client with given options/credentials
func GetClient(cfg config.Jenkins) (*gojenkins.Jenkins, error) {
	if !cfg.IsEnabled() {
		return nil, nil
	}

	jenkinsClient := gojenkins.CreateJenkins(
		client.GetHTTPClient(),
		cfg.Host,
		cfg.Username,
		cfg.Password,
	)

	return jenkinsClient.Init(context.TODO())
}
