package config

type Aws struct {
	Enabled bool     `mapstructure:"enabled"`
	Alias   string   `mapstructure:"alias,omitempty"`
	Lambda  []Lambda `mapstructure:"lambdas"`
}
type Lambda struct {
	Name     string  `mapstructure:"name"`
	Desc     string  `mapstructure:"desc"`
	FuncName string  `mapstructure:"funcName"`
	Inputs   []Input `mapstructure:"inputs"`
}
type Input struct {
	Key  string `mapstructure:"key"`
	Desc string `mapstructure:"desc"`
}

func (c Aws) IsEnabled() bool {
	return c.Enabled
}
