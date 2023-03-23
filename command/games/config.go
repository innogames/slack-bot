package games

import (
	"github.com/innogames/slack-bot/v2/bot/config"
)

// Config to enable/disable games
type Config struct {
	Enabled bool `mapstructure:"enabled"`
}

func loadConfig(config *config.Config) Config {
	cfg := Config{}
	cfg.Enabled = true
	_ = config.LoadCustom("games", &cfg)

	return cfg
}
