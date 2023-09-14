package config

import "time"

type VCS struct {
	Type           string        `mapstructure:"type"` // stash/bitbucket/git/null
	Repository     string        `mapstructure:"repository"`
	UpdateInterval time.Duration `mapstructure:"update_interval"`
}

func (c VCS) IsEnabled() bool {
	return c.Type != "" && c.UpdateInterval > 0
}
