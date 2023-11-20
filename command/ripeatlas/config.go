package ripeatlas

import (
	"time"

	"github.com/innogames/slack-bot/v2/bot/config"
)

// Config configuration: API key to do API calls
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

func loadConfig(config *config.Config) Config {
	cfg := defaultConfig
	_ = config.LoadCustom("ripeatlas", &cfg)

	return cfg
}
