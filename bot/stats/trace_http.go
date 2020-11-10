package stats

import (
	"net/http/httptrace"
)

func GetHttpTracer() httptrace.ClientTrace {
	return httptrace.ClientTrace{
		GotConn: func(connInfo httptrace.GotConnInfo) {
			IncreaseOne("http_request")
		},
	}
}
