package jenkins

//go:generate $GOPATH/bin/mockery --output ../../mocks -name Client
//go:generate $GOPATH/bin/mockery --output ../../mocks -name Job

import (
	"github.com/bndr/gojenkins"
	"github.com/innogames/slack-bot/config"

	"net/http"
)

func GetClient(cfg config.Jenkins) (*gojenkins.Jenkins, error) {
	if !cfg.IsEnabled() {
		return nil, nil
	}

	jenkinsClient := gojenkins.CreateJenkins(
		&http.Client{},
		cfg.Host,
		cfg.Username,
		cfg.Password,
	)

	return jenkinsClient.Init()
}

// Client is a interface of gojenkins.Jenkins
type Client interface {
	GetJob(id string, parentIDs ...string) (*gojenkins.Job, error)
	BuildJob(name string, options ...interface{}) (int64, error)
	GetQueue() (*gojenkins.Queue, error)
	GetAllNodes() ([]*gojenkins.Node, error)
}
