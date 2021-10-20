package jenkins

//go:generate $GOPATH/bin/mockery --output ../../mocks --name Client

import (
	"context"

	"github.com/bndr/gojenkins"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/client"
)

// Client is an interface representing used jenkins functions of gojenkins.
type Client interface {
	GetJob(ctx context.Context, id string) (*gojenkins.Job, error)
	BuildJob(ctx context.Context, name string, params map[string]string) (int64, error)
	GetAllNodes(ctx context.Context) ([]*gojenkins.Node, error)
}

// GetClient created Jenkins client with given options/credentials
func GetClient(cfg config.Jenkins) (Client, error) {
	if !cfg.IsEnabled() {
		return nil, nil
	}

	return createJenkinsClient(context.TODO(), client.GetHTTPClient(), cfg)
}
