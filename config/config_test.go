package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadExampleConfig(t *testing.T) {
	cfg, err := LoadPattern("../example_config/*.yaml")
	assert.Nil(t, err)
	assert.NotEmpty(t, cfg.Slack)
	assert.NotEmpty(t, cfg.Macros)
}

func TestLoadNotMatchingPattern(t *testing.T) {
	cfg, err := LoadPattern("notexistingconfig.yaml")
	assert.EqualError(t, err, "no config file found: notexistingconfig.yaml")
	assert.Equal(t, defaultConfig, cfg)
}

func TestInvalidFile(t *testing.T) {
	cfg, err := LoadConfig("../fooo.yaml")
	assert.EqualError(t, err, "failed to load config file from ../fooo.yaml: open ../fooo.yaml: no such file or directory")
	assert.Equal(t, defaultConfig, cfg)
}
