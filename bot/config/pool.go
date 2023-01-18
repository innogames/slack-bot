package config

import "time"

// Pool config contains the Resources of the Pool
type Pool struct {
	LockDuration time.Duration
	NotifyExpire time.Duration
	Resources    []*Resource
}

// Resource config contains definitions about the
type Resource struct {
	Name         string
	ExplicitLock bool
	Addresses    []string
	Features     []string
}

// IsEnabled checks if there are resources in the pool
func (c *Pool) IsEnabled() bool {
	return len(c.Resources) > 0
}
