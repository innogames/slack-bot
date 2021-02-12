package queue

import (
	"sync"
)

// list of currently running commands
var runningCommands = map[string]*RunningCommand{}

// RunningCommand is a wrapper to sync.WaitGroup to control the behavior of a running command:
// - when the command is done, call the Done() method
// - listener can register via Wait() method which is blocking until command is done
type RunningCommand struct {
	wg sync.WaitGroup
}

// Wait blocks until command is Done()
func (r *RunningCommand) Wait() {
	r.wg.Wait()
}

// Done will finish the sync.WaitGroup and releases the lock for Done()
func (r *RunningCommand) Done() {
	r.wg.Done()
}
