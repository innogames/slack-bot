package ripe_atlas

import (
	"github.com/innogames/slack-bot/v2/bot/config"
)

// Config configuration: API key to do API calls
type Config struct {
	APIKey string `mapstructure:"api_key"`
	APIURL string `mapstructure:"api_host"`
}

// IsEnabled checks if token is set
func (c *Config) IsEnabled() bool {
	return c.APIKey != ""
}

var defaultConfig = Config{
	APIURL: "https://atlas.ripe.net/api/v2",
}

func loadConfig(config *config.Config) Config {
	cfg := defaultConfig
	_ = config.LoadCustom("ripe_atlas", &cfg)

	return cfg
}
