package util

import (
	"io/ioutil"
	"net/http"
	"time"
)

// DoGetRequest will a a pretty simple GET request to fetch content via HTTP
func DoGetRequest(url string) ([]byte, error) {
	// Build an http client so we can have control over timeout
	client := &http.Client{
		Timeout: time.Second * 2,
	}

	res, getErr := client.Get(url)
	if getErr != nil {
		return nil, getErr
	}

	// defer the closing of the res body
	defer res.Body.Close()

	// read the http response body into a byte stream
	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return nil, readErr
	}

	return body, nil
}
