package client

import (
	"net/http"
	"time"
)

// HTTPClient default http client for this bot to use the default go-client with a Timeout
var HTTPClient = &http.Client{
	Timeout: time.Second * 10,
}
