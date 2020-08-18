package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadExampleConfig(t *testing.T) {
	cfg, err := LoadPattern("../../config.example.yaml")
	assert.Nil(t, err)
	assert.NotNil(t, cfg.Slack)
	assert.NotEmpty(t, cfg.Macros)

	assert.Equal(t, false, cfg.Jenkins.IsEnabled())
	assert.Equal(t, false, cfg.Jira.IsEnabled())
	assert.Equal(t, false, cfg.Mqtt.IsEnabled())
	assert.Equal(t, false, cfg.Bitbucket.IsEnabled())
	assert.Equal(t, false, cfg.Server.IsEnabled())
}

func TestLoadNotMatchingPattern(t *testing.T) {
	cfg, err := LoadPattern("notexistingconfig.yaml")
	assert.EqualError(t, err, "no config file found: notexistingconfig.yaml")
	assert.Equal(t, defaultConfig, cfg)
}

func TestInvalidFiles(t *testing.T) {
	cfg, err := LoadPattern("../neneneee*yaml")
	assert.EqualError(t, err, "no config file found: ../neneneee*yaml")
	assert.Equal(t, defaultConfig, cfg)

	cfg, err = loadConfig("../fooo.yaml")
	assert.EqualError(t, err, "failed to load config file from ../fooo.yaml: open ../fooo.yaml: no such file or directory")
	assert.Equal(t, Config{}, cfg)

	cfg, err = loadConfig("../../Makefile")
	assert.EqualError(t, err, "failed to parse configuration file: yaml: line 7: found character that cannot start any token")
	assert.Equal(t, Config{}, cfg)
}
