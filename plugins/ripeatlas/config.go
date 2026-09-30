package ripeatlas

import (
	"time"
)

// Config of the ripeatlas plugin, defined in "plugins.ripeatlas.config". The API key is needed to do API calls
type Config struct {
	APIKey         string        `mapstructure:"api_key"`
	APIURL         string        `mapstructure:"api_url"`
	StreamURL      string        `mapstructure:"stream_url"`
	UpdateInterval time.Duration `mapstructure:"update_interval"`
}

// IsEnabled checks if token is set
func (c *Config) IsEnabled() bool {
	return c.APIKey != ""
}

var defaultConfig = Config{
	APIURL:         "https://atlas.ripe.net/api/v2",
	StreamURL:      "https://atlas-stream.ripe.net/stream/",
	UpdateInterval: time.Second,
}
