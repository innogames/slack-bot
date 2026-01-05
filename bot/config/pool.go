package config

import "time"

// Pool config contains the Resources of the Pool
type Pool struct {
	LockDuration time.Duration `yaml:"lock_duration"`
	NotifyExpire time.Duration `yaml:"notify_expire"`
	Resources    []*Resource   `yaml:"resources"`
}

// Resource config contains definitions about the
type Resource struct {
	Name         string   `yaml:"name"`
	ExplicitLock bool     `yaml:"explicit_lock"`
	Addresses    []string `yaml:"addresses"`
	Features     []string `yaml:"features"`
}

// IsEnabled checks if there are resources in the pool
func (c *Pool) IsEnabled() bool {
	return len(c.Resources) > 0
}
