package jenkins

//go:generate $GOPATH/bin/mockery --output ../../mocks -name Client
//go:generate $GOPATH/bin/mockery --output ../../mocks -name Job

import (
	"github.com/bndr/gojenkins"
	"github.com/innogames/slack-bot/bot/config"
)

// Client is a interface of gojenkins.Jenkins
type Client interface {
	GetJob(id string, parentIDs ...string) (*gojenkins.Job, error)
	BuildJob(name string, options ...interface{}) (int64, error)
	GetQueue() (*gojenkins.Queue, error)
	GetAllNodes() ([]*gojenkins.Node, error)
}

// GetClient created Jenkins client with given options/credentials
func GetClient(cfg config.Jenkins) (*gojenkins.Jenkins, error) {
	if !cfg.IsEnabled() {
		return nil, nil
	}

	jenkinsClient := gojenkins.CreateJenkins(
		nil,
		cfg.Host,
		cfg.Username,
		cfg.Password,
	)

	return jenkinsClient.Init()
}
