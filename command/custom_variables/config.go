package custom_variables

import "github.com/innogames/slack-bot/v2/bot/config"

// Config to enable/disable custom variables
type Config struct {
	Enabled bool `mapstructure:"enabled"`
}

func loadConfig(config *config.Config) Config {
	cfg := Config{}
	cfg.Enabled = true
	_ = config.LoadCustom("custom_variables", &cfg)

	return cfg
}
