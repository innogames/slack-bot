package client

import (
	"fmt"
	"net/http"
	"time"

	"github.com/innogames/slack-bot.v2/bot/version"
)

// GetHTTPClient returns a default http client for this bot to use the default go-client with a Timeout
func GetHTTPClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxConnsPerHost = 3
	transport.MaxIdleConns = 5
	transport.IdleConnTimeout = time.Second * 15

	return &http.Client{
		Timeout:   time.Second * 10,
		Transport: &setUserAgentHeader{transport},
	}
}

type setUserAgentHeader struct {
	roundTripper http.RoundTripper
}

func (t *setUserAgentHeader) RoundTrip(req *http.Request) (*http.Response, error) {
	userAgent := fmt.Sprintf("slack-bot/%s", version.Version)
	req.Header.Add("User-Agent", userAgent)

	return t.roundTripper.RoundTrip(req)
}
