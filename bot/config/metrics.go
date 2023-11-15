package config

type Metrics struct {
	// e.g. use ":8082" to expose metrics on all interfaces
	PrometheusListener string `mapstructure:"prometheus_listener"`
}

// IsEnabled returns true if the metrics are enabled by config
func (c *Metrics) IsEnabled() bool {
	return c.PrometheusListener != ""
}
