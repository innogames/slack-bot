package openai

import "github.com/innogames/slack-bot/v2/bot/config"

// Config configuration: API key to do API calls
type Config struct {
	APIKey               string  `mapstructure:"api_key"`
	APIHost              string  `mapstructure:"api_host"`
	InitialSystemMessage string  `mapstructure:"initial_system_message"`
	Model                string  `mapstructure:"host"`
	Temperature          float32 `mapstructure:"temperature"`
}

// IsEnabled checks if token is set
func (c *Config) IsEnabled() bool {
	return c.APIKey != ""
}

var defaultConfig = Config{
	APIHost: apiHost,
	Model:   defaultModel,
}

func loadConfig(config *config.Config) Config {
	cfg := defaultConfig
	config.LoadCustom("openai", &cfg)

	return cfg
}
