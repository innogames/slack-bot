package config

type Aws struct {
	Enabled bool     `mapstructure:"enabled"`
	Lambda  []Lambda `mapstructure:"lambda"`
}
type Lambda struct {
	Name        string `mapstructure:"name"`
	Alias       string `mapstructure:"alias,omitempty"`
	Description string `mapstructure:"description"`
}

func (c Aws) IsEnabled() bool {
	return c.Enabled
}
