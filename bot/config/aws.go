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
	Description string   `mapstructure:"description"`
}
type LambdaReturnCode struct {
	Code string `json:"code"`
}
type LambdaSuccessMessage struct {
	Message []map[string]interface{} `json:"message"`
}

type LambdaFailedMessage struct {
	Message string `json:"message"`
}

func (c Aws) IsEnabled() bool {
	return c.Enabled
}
