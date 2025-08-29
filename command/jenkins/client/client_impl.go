package client

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/bndr/gojenkins"
	"github.com/innogames/slack-bot/v2/bot/config"
)

// implementation of Client interface. proxies to gojenkins with additional handling for inner jenkins jobs.
type jenkinsClientImpl struct {
	client *gojenkins.Jenkins
}

func createJenkinsClient(ctx context.Context, httpClient *http.Client, cfg config.Jenkins) (*jenkinsClientImpl, error) {
	var jenkins *gojenkins.Jenkins
	if cfg.Username == "" {
		jenkins = gojenkins.CreateJenkins(
			httpClient,
			cfg.Host,
		)
	} else {
		jenkins = gojenkins.CreateJenkins(
			httpClient,
			cfg.Host,
			cfg.Username,
			cfg.Password,
		)
	}

	jenkinsClient, err := jenkins.Init(ctx)
	if err != nil {
		return nil, err
	}

	client := &jenkinsClientImpl{
		client: jenkinsClient,
	}

	return client, nil
}

func (c *jenkinsClientImpl) GetJob(ctx context.Context, id string) (*gojenkins.Job, error) {
	// First decode any URL-encoded characters in the job ID
	decodedID := id
	if strings.Contains(id, "%") {
		if decoded, err := url.QueryUnescape(id); err == nil {
			decodedID = decoded
		}
	}

	// split jobs id by "/" to be able to access inner job
	jobs := strings.Split(decodedID, "/")

	jobsCount := len(jobs)
	if jobsCount > 1 {
		return c.client.GetJob(ctx, jobs[jobsCount-1], jobs[:jobsCount-1]...)
	}

	return c.client.GetJob(ctx, decodedID)
}

func (c *jenkinsClientImpl) BuildJob(ctx context.Context, name string, params map[string]string) (int64, error) {
	return c.client.BuildJob(ctx, name, params)
}

func (c *jenkinsClientImpl) GetAllNodes(ctx context.Context) ([]*gojenkins.Node, error) {
	return c.client.GetAllNodes(ctx)
}
