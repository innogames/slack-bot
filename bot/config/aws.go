package config

type Aws struct {
	CloudFront []AwsCfDistribution `mapstructure:"cloud_front"`
}

type AwsCfDistribution struct {
	Id   string `mapstructure:"id"`
	Name string `mapstructure:"name"`
}
