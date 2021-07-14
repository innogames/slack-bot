package client

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHttp(t *testing.T) {
	t.Run("test user agent", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			assert.Equal(t, req.Header.Get("User-Agent"), "slack-bot/unknown")
			rw.Write([]byte("ok"))
		}))
		defer server.Close()

		client := GetHTTPClient()
		resp, err := client.Get(server.URL)
		assert.Nil(t, err)
		defer resp.Body.Close()

		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		assert.Equal(t, []byte("ok"), bodyBytes)
	})
}
