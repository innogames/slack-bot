package interaction

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/innogames/slack-bot/bot/config"
	"github.com/sirupsen/logrus/hooks/test"
)

func TestHealthCheckHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	server := Server{}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.healthCheckHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

// todo catching just some flows yet!
func TestHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/commands", nil)
	if err != nil {
		t.Fatal(err)
	}

	logger, _ := test.NewNullLogger()
	server := Server{}
	server.logger = logger
	server.cfg = config.Server{
		Listen:             "0.0.0.0:80",
		VerificationSecret: "iamsecret",
	}

	rr := httptest.NewRecorder()

	req.Header.Set("X-Slack-Signature", "v0=a2114d57b48eac39b9ad189dd8316235a7b4a8d21a10bd27519666489c69b503")
	req.Header.Set("X-Slack-Request-Timestamp", "1531420618")
	handler := http.HandlerFunc(server.interactionHandler)
	handler.ServeHTTP(rr, req)

	// Check timestamp too old
	expected := "Could not initialize SecretVerifier: timestamp is too old"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
