package config

// Cron is represents a single cron which can be configured
type Cron struct {
	Channel  string
	Schedule string
	Commands []string
}
