package aws

// Config of the aws plugin, read from the "plugins: aws:" config section
// (the legacy top-level "aws:" key is still supported)
type Config struct {
	Enabled    bool             `mapstructure:"enabled"`
	CloudFront []CfDistribution `mapstructure:"cloud_front"`
}

// CfDistribution represents a CloudFront distribution with a human readable name
type CfDistribution struct {
	ID   string `mapstructure:"id"`
	Name string `mapstructure:"name"`
}

// IsEnabled checks if the aws plugin is activated in the config
func (c Config) IsEnabled() bool {
	return c.Enabled
}
