package aws

// Config of the aws plugin, defined in "plugins.aws.config"
type Config struct {
	CloudFront []CfDistribution `mapstructure:"cloud_front"`
}

// CfDistribution is a CloudFront distribution with a human readable name
type CfDistribution struct {
	ID   string `mapstructure:"id"`
	Name string `mapstructure:"name"`
}
