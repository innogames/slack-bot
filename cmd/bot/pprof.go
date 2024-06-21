//go:build pprof
// +build pprof

package main

// build tag to enable pprof server
//
// https://golang.org/doc/diagnostics.html
//
// curl localhost:6968/debug/pprof/heap?debug=1 | less
// curl localhost:6968/debug/pprof/allocs?debug=1 | less
// curl localhost:6968/debug/pprof/goroutine?debug=1 | less
// curl localhost:6968/debug/pprof/profile?seconds=10 > /tmp/profile.pprof # -> go tool pprof profile.pprof
// curl localhost:6968/debug/pprof/trace?seconds=10 > /tmp/trace.pprof # -> go tool trace trace.pprof

import (
	"net/http"
	_ "net/http/pprof"

	log "github.com/sirupsen/logrus"
)

func init() {
	go func() {
		log.Info("Starting pprof server on http://localhost:6968/debug/pprof/")
		log.Info(http.ListenAndServe("localhost:6968", nil))
	}()
}
