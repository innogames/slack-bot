package openai

import (
	"time"

	"github.com/innogames/slack-bot/v2/bot/config"
)

// Config configuration: API key to do API calls
type Config struct {
	APIKey               string  `mapstructure:"api_key"`
	APIHost              string  `mapstructure:"api_host"`
	InitialSystemMessage string  `mapstructure:"initial_system_message"`
	Model                string  `mapstructure:"model"`
	Temperature          float32 `mapstructure:"temperature"`
	Seed                 string  `mapstructure:"seed"`
	MaxTokens            int     `mapstructure:"max_tokens"`

	// number of thread messages stored which are used as a context for further requests
	HistorySize int `mapstructure:"history_size"`

	// is no other command matched, evaluate the message with openai
	UseAsFallback bool `mapstructure:"use_as_fallback"`

	// maximum update frequency of slack messages when "stream" is active
	UpdateInterval time.Duration `mapstructure:"update_interval"`

	// log all input+output text to the logger. This could include personal information, therefore disabled by default!
	LogTexts bool `mapstructure:"log_texts"`
}

// IsEnabled checks if token is set
func (c *Config) IsEnabled() bool {
	return c.APIKey != ""
}

var defaultConfig = Config{
	APIHost:              apiHost,
	Model:                "gpt-3.5-turbo", // aka model behind ChatGPT
	UpdateInterval:       time.Second,
	HistorySize:          15,
	InitialSystemMessage: "You are a helpful Slack bot. By default, keep your answer short and truthful",
}

func loadConfig(config *config.Config) Config {
	cfg := defaultConfig
	_ = config.LoadCustom("openai", &cfg)

	return cfg
}
