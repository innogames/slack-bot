package config

type Aws struct {
	Enabled bool     `mapstructure:"enabled"`
	Lambda  []Lambda `mapstructure:"lambda"`
}
type Lambda struct {
	Name        string   `mapstructure:"name"`
	Alias       string   `mapstructure:"alias,omitempty"`
	Inputs      []string `mapstructure:"inputs,omitempty"`
	Outputs     []string `mapstructure:"outputs"`
	Description string   `mapstructure:"description,omitempty"`
}
type LambdaOutput struct {
	Code    string                   `json:"code"`
	Message []map[string]interface{} `json:"message,omitempty"`
	Error   string                   `json:"error,omitempty"`
}

func (c Aws) IsEnabled() bool {
	return c.Enabled
}
