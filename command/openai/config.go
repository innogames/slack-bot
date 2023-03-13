package openai

import (
	"time"

	"github.com/innogames/slack-bot/v2/bot/config"
)

// Config configuration: API key to do API calls
type Config struct {
	APIKey               string        `mapstructure:"api_key"`
	APIHost              string        `mapstructure:"api_host"`
	InitialSystemMessage string        `mapstructure:"initial_system_message"`
	Model                string        `mapstructure:"model"`
	Temperature          float32       `mapstructure:"temperature"`
	UpdateInterval       time.Duration `mapstructure:"update_interval"`
}

// IsEnabled checks if token is set
func (c *Config) IsEnabled() bool {
	return c.APIKey != ""
}

var defaultConfig = Config{
	APIHost:              apiHost,
	Model:                defaultModel,
	UpdateInterval:       time.Second,
	InitialSystemMessage: "You are a helpful Slack bot. By default, keep your answer short and truthful",
}

func loadConfig(config *config.Config) Config {
	cfg := defaultConfig
	_ = config.LoadCustom("openai", &cfg)

	return cfg
}
