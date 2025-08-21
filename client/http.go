package client

import (
	"net/http"
	"sync"
	"time"

	"github.com/innogames/slack-bot/v2/bot/version"
)

var (
	httpClient    *http.Client
	getHTTPClient sync.Once
)

// GetHTTPClient returns the http client for this bot to use the default go-client with a Timeout of 10s
func GetHTTPClient() *http.Client {
	getHTTPClient.Do(func() {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.MaxConnsPerHost = 3

		httpClient = &http.Client{
			Timeout:   time.Second * 15,
			Transport: &botTransport{transport},
		}
	})

	return httpClient
}

// custom http.Transport to set a custom user-agent
type botTransport struct {
	roundTripper http.RoundTripper
}

// RoundTrip add the User-Agent header containing the slack-bot version to identify traffic from this bot
func (t *botTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	userAgent := "slack-bot/" + version.Version
	req.Header.Add("User-Agent", userAgent)

	return t.roundTripper.RoundTrip(req)
}
