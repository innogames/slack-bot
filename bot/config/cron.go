package config

// Cron is represents a single cron which can be configured
type Cron struct {
	Channel  string   `mapstructure:"channel"`
	Schedule string   `mapstructure:"schedule"`
	Commands []string `mapstructure:"commands"`
}
