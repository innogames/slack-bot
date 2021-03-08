package client

import (
	"net/http"
	"time"
)

// GetHTTPClient returns a default http client for this bot to use the default go-client with a Timeout
func GetHTTPClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxConnsPerHost = 3
	transport.MaxIdleConns = 5
	transport.IdleConnTimeout = time.Second * 15

	return &http.Client{
		Timeout:   time.Second * 10,
		Transport: transport,
	}
}
