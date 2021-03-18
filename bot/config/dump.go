package config

import (
	"encoding/json"
)

// Dump the given config in a json string
func Dump(cfg Config) string {
	bytes, _ := json.MarshalIndent(cfg, "", "   ")

	return string(bytes)
}
