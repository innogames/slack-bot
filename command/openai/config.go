package openai

// Config configuration: API key to do API calls
type Config struct {
	APIKey string `mapstructure:"api_key"`
	// todo add more config, like temperature, max token etc
}

// IsEnabled checks if token is set
func (c *Config) IsEnabled() bool {
	return c.APIKey != ""
}
