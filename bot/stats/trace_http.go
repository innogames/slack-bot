package stats

import (
	"net/http/httptrace"
)

func GetHTTPTracer() httptrace.ClientTrace {
	return httptrace.ClientTrace{
		GotConn: func(connInfo httptrace.GotConnInfo) {
			IncreaseOne("http_request")
		},
	}
}
