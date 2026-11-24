package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bndr/gojenkins"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetJobRequestPath(t *testing.T) {
	tests := []struct {
		name         string
		jobName      string
		expectedPath string
	}{
		{
			name:         "root job",
			jobName:      "deployer",
			expectedPath: "/job/deployer/api/json",
		},
		{
			name:         "multibranch main child",
			jobName:      "deployer/main",
			expectedPath: "/job/deployer/job/main/api/json",
		},
		{
			name:         "multibranch pull request child",
			jobName:      "deployer/PR-123",
			expectedPath: "/job/deployer/job/PR-123/api/json",
		},
		{
			name:         "branch name with encoded slash",
			jobName:      "deployer/feature%2Ffoo",
			expectedPath: "/job/deployer/job/feature%2Ffoo/api/json",
		},
		{
			name:         "branch name with encoded special characters",
			jobName:      "folder/deployer/bugfix%2FABC-123%2Btest",
			expectedPath: "/job/folder/job/deployer/job/bugfix%2FABC-123%2Btest/api/json",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var actualPath string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				actualPath = r.URL.EscapedPath()
				w.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprint(w, `{}`)
			}))
			defer server.Close()

			client := &jenkinsClientImpl{
				client: gojenkins.CreateJenkins(server.Client(), server.URL),
			}

			_, err := client.GetJob(context.Background(), test.jobName)
			require.NoError(t, err)
			assert.Equal(t, test.expectedPath, actualPath)
		})
	}
}

func TestBuildMultibranchJob(t *testing.T) {
	var requests []string
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, fmt.Sprintf("%s %s", r.Method, r.URL.EscapedPath()))

		if r.Method == http.MethodPost {
			assert.NoError(t, r.ParseForm())
			assert.Equal(t, "bar", r.Form.Get("foo"))
			w.Header().Set("Location", server.URL+"/queue/item/42/")
			w.WriteHeader(http.StatusCreated)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"property":[{"parameterDefinitions":[{"name":"foo"}]}]}`)
	}))
	defer server.Close()

	client := &jenkinsClientImpl{
		client: gojenkins.CreateJenkins(server.Client(), server.URL),
	}

	queueID, err := client.BuildJob(
		context.Background(),
		"deployer/feature%2Ffoo",
		map[string]string{"foo": "bar"},
	)
	require.NoError(t, err)
	assert.Equal(t, int64(42), queueID)
	assert.Equal(t, []string{
		"GET /job/deployer/job/feature%2Ffoo/api/json",
		"GET /job/deployer/job/feature%2Ffoo/api/json",
		"GET /job/deployer/job/feature%2Ffoo/api/json",
		"GET /crumbIssuer/api/json/api/json",
		"POST /job/deployer/job/feature%2Ffoo/buildWithParameters",
	}, requests)
}
