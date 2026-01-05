package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bndr/gojenkins"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		require.NoError(t, err)

		stop := make(chan bool, 1)
		builds, err := WatchJob(ctx, client, "notExistingJob", stop)

		assert.Equal(t, errors.New("404-fail"), err)
		assert.Nil(t, builds)
	})

	t.Run("Watch Job", func(t *testing.T) {
		ctx := context.Background()
		client, err := GetClient(cfg)
		require.NoError(t, err)

		stop := make(chan bool, 1)
		builds, err := WatchJob(ctx, client, "testJob", stop)

		require.NoError(t, err)
		assert.Empty(t, builds)

		stop <- true
		assert.Empty(t, builds)
	})

	t.Run("Watch Job with invalid build", func(t *testing.T) {
		ctx := context.Background()
		client, err := GetClient(cfg)
		require.NoError(t, err)

		stop := make(chan bool, 1)
		builds, err := WatchJob(ctx, client, "testJob2", stop)

		assert.Equal(t, errors.New("404-fail"), err)
		assert.Empty(t, builds)
	})
}

func spawnJobWatcherServer() *httptest.Server {
	mux := http.NewServeMux()

	// test connection - return valid JSON for gojenkins Init
	mux.HandleFunc("/api/json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"_class":"hudson.model.Hudson"}`))
	})

	mux.HandleFunc("/job/testJob/api/json", func(w http.ResponseWriter, _ *http.Request) {
		job := gojenkins.JobResponse{}
		job.Name = "test"
		job.LastBuild = gojenkins.JobBuild{
			Number: 42,
		}
		encoder := json.NewEncoder(w)
		encoder.Encode(job)
	})

	mux.HandleFunc("/job/testJob2/api/json", func(w http.ResponseWriter, _ *http.Request) {
		job := gojenkins.JobResponse{}
		job.Name = "test"
		job.LastBuild = gojenkins.JobBuild{
			Number: 42,
		}
		encoder := json.NewEncoder(w)
		encoder.Encode(job)
	})

	mux.HandleFunc("/job/notExistingJob/api/json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add("X-Error", "404-fail")
		w.WriteHeader(404)
	})

	mux.HandleFunc("/job/testJob/42/api/json", func(w http.ResponseWriter, _ *http.Request) {
		build := gojenkins.BuildResponse{}
		build.Number = 42
		build.Building = true

		encoder := json.NewEncoder(w)
		encoder.Encode(build)
	})

	mux.HandleFunc("/job/testJob2/42/api/json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add("X-Error", "404-fail")
		w.WriteHeader(404)
	})

	return httptest.NewServer(mux)
}
