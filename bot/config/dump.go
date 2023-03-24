package config

import (
	"gopkg.in/yaml.v3"
)

// Dump the given config in a yaml string
func Dump(cfg Config) string {
	bytes, _ := yaml.Marshal(cfg)

	return string(bytes)
}
