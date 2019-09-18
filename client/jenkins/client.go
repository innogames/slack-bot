package jenkins

import (
	"github.com/brainexe/gojenkins"
	"github.com/innogames/slack-bot/bot/config"
)

// Client is a interface of gojenkins.Jenkins
type Client interface {
	GetJob(id string, parentIDs ...string) (*gojenkins.Job, error)
	BuildJob(name string, options ...interface{}) (int64, error)
	GetQueue() (*gojenkins.Queue, error)
	GetAllNodes() ([]*gojenkins.Node, error)
}

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
