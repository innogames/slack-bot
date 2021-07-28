package jenkins

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bndr/gojenkins"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/stretchr/testify/assert"
)

func TestWatchJob(t *testing.T) {
	server := spawnJobWatcherServer()
	defer server.Close()

	cfg := config.Jenkins{
		Host: server.URL,
	}

	t.Run("Watch not existing Job", func(t *testing.T) {
		ctx := context.Background()
		client, err := GetClient(cfg)
		assert.Nil(t, err)

		stop := make(chan bool, 1)
		builds, err := WatchJob(ctx, client, "notExistingJob", stop)

		assert.Equal(t, fmt.Errorf("404-fail"), err)
		assert.Nil(t, builds)
	})

	t.Run("Watch Job", func(t *testing.T) {
		ctx := context.Background()
		client, err := GetClient(cfg)
		assert.Nil(t, err)

		stop := make(chan bool, 1)
		builds, err := WatchJob(ctx, client, "testJob", stop)

		assert.Nil(t, err)
		assert.Len(t, builds, 0)

		stop <- true
		assert.Len(t, builds, 0)
	})

	t.Run("Watch Job with invalid build", func(t *testing.T) {
		ctx := context.Background()
		client, err := GetClient(cfg)
		assert.Nil(t, err)

		stop := make(chan bool, 1)
		builds, err := WatchJob(ctx, client, "testJob2", stop)

		assert.Equal(t, fmt.Errorf("404-fail"), err)
		assert.Len(t, builds, 0)
	})
}

func spawnJobWatcherServer() *httptest.Server {
	mux := http.NewServeMux()

	// test connection
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`ok`))
	})

	mux.HandleFunc("/job/testJob/api/json", func(w http.ResponseWriter, r *http.Request) {
		job := gojenkins.JobResponse{}
		job.Name = "test"
		job.LastBuild = gojenkins.JobBuild{
			Number: 42,
		}
		encoder := json.NewEncoder(w)
		encoder.Encode(job)
	})

	mux.HandleFunc("/job/testJob2/api/json", func(w http.ResponseWriter, r *http.Request) {
		job := gojenkins.JobResponse{}
		job.Name = "test"
		job.LastBuild = gojenkins.JobBuild{
			Number: 42,
		}
		encoder := json.NewEncoder(w)
		encoder.Encode(job)
	})

	mux.HandleFunc("/job/notExistingJob/api/json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-Error", "404-fail")
		w.WriteHeader(404)
	})

	mux.HandleFunc("/job/testJob/42/api/json", func(w http.ResponseWriter, r *http.Request) {
		build := gojenkins.BuildResponse{}
		build.Number = 42
		build.Building = true

		encoder := json.NewEncoder(w)
		encoder.Encode(build)
	})

	mux.HandleFunc("/job/testJob2/42/api/json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-Error", "404-fail")
		w.WriteHeader(404)
	})

	return httptest.NewServer(mux)
}
