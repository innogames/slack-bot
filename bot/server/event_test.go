package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/innogames/slack-bot/bot/config"
	"github.com/stretchr/testify/assert"
)

func TestEvent(t *testing.T) {
	server := Server{}
	server.cfg = config.Server{
		Listen:        "0.0.0.0:80",
		SigningSecret: "iamsecret",
	}

	rr := httptest.NewRecorder()

	t.Run("Test expired event", func(t *testing.T) {
		body := strings.NewReader("...")
		req, _ := http.NewRequest("GET", "/events", body)
		req.Header.Set("X-Slack-Signature", "v0=a2114d57b48eac39b9ad189dd8316235a7b4a8d21a10bd27519666489c69b503")
		req.Header.Set("X-Slack-Request-Timestamp", "1531420618")

		handler := http.HandlerFunc(server.eventHandler)
		handler.ServeHTTP(rr, req)

		// Check timestamp too old
		expected := "timestamp is too old"
		if rr.Body.String() != expected {
			t.Errorf("handler returned unexpected body: got %v want %v",
				rr.Body.String(), expected)
		}
	})

	t.Run("Test invalid signature event", func(t *testing.T) {
		body := strings.NewReader("...")
		req, _ := http.NewRequest("GET", "/events", body)
		req.Header.Set("X-Slack-Signature", "v0=73da409a5cbd4633aea2f00baf1a8571bef26f553bca8a1ec75ccb4aef859f32")
		req.Header.Set("X-Slack-Request-Timestamp", fmt.Sprint(time.Now().Unix()))

		handler := http.HandlerFunc(server.eventHandler)
		handler.ServeHTTP(rr, req)

		assert.Contains(t, rr.Body.String(), "Computed unexpected signature of")
	})
}
