package jenkins

//go:generate $GOPATH/bin/mockery --output ../../mocks --name Client

import (
	"github.com/bndr/gojenkins"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/client"
)

// Client is a interface of gojenkins.Jenkins
type Client interface {
	GetJob(id string, parentIDs ...string) (*gojenkins.Job, error)
	BuildJob(name string, options ...interface{}) (int64, error)
	GetAllNodes() ([]*gojenkins.Node, error)
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

	return jenkinsClient.Init()
}
