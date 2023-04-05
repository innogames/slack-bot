package custom_commmands

import "github.com/innogames/slack-bot/v2/bot/config"

// Config to enable/disable custom commands
type Config struct {
	Enabled bool `mapstructure:"enabled"`
}

func loadConfig(config *config.Config) Config {
	cfg := Config{}
	cfg.Enabled = true
	_ = config.LoadCustom("custom_commands", &cfg)

	return cfg
}
