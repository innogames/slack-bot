package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHttp(t *testing.T) {
	t.Run("test user agent", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			assert.Equal(t, "slack-bot/unknown", req.Header.Get("User-Agent"))
			rw.Write([]byte("ok"))
		}))
		defer server.Close()

		client := GetHTTPClient()
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
		require.NoError(t, err)
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Equal(t, []byte("ok"), bodyBytes)
	})
}
