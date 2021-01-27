package util

import (
	"context"
	"sync"

	log "github.com/sirupsen/logrus"
)

// NewServerContext wrapper for ctx to simply add childs which have a blocking shutdown process
// -> make sure all stuff is closed properly before exit
// todo directly add term/kill handling
func NewServerContext() *ServerContext {
	ctx, cancel := context.WithCancel(context.Background())

	return &ServerContext{
		Context: ctx,
		wg:      &sync.WaitGroup{},
		cancel:  cancel,
	}
}

type ServerContext struct {
	context.Context

	wg     *sync.WaitGroup
	cancel context.CancelFunc
}

// StopTheWorld start the shutdown process
func (c *ServerContext) StopTheWorld() {
	log.Info("Stop the world!")
	c.cancel()
	log.Info("Waiting for subcommands to finish...")
	c.wg.Wait()
	log.Info("Done...bye bye!")
}

// add a new child...
func (c *ServerContext) RegisterChild() {
	c.wg.Add(1)
}

// ...mark a child shutdown as done (-> use this method in defer)
func (c *ServerContext) ChildDone() {
	c.wg.Done()
}
