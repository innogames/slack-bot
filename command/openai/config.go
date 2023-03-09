package openai

// OpenAIConfig configuration: API key to do API calls
type OpenAIConfig struct {
	ApiKey string `mapstructure:"api_key"`
}

// IsEnabled checks if toekn is set
func (c *OpenAIConfig) IsEnabled() bool {
	return c.ApiKey != ""
}
