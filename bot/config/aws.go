package config

type Aws struct {
	Enabled    bool                `mapstructure:"enabled"`
	CloudFront []AwsCfDistribution `mapstructure:"cloud_front"`
}

type AwsCfDistribution struct {
	ID   string `mapstructure:"id"`
	Name string `mapstructure:"name"`
}

func (c Aws) IsEnabled() bool {
	return c.Enabled
}
