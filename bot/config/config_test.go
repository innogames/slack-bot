package config

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

func TestLoadExampleConfig(t *testing.T) {
	cfg, err := Load("../../config.example.yaml")
	assert.Nil(t, err)
	assert.NotNil(t, cfg.Slack)
	assert.NotEmpty(t, cfg.Commands)

	assert.Equal(t, false, cfg.Jenkins.IsEnabled())
	assert.Equal(t, false, cfg.Jira.IsEnabled())
	assert.Equal(t, false, cfg.Mqtt.IsEnabled())
	assert.Equal(t, false, cfg.Bitbucket.IsEnabled())
	assert.Equal(t, false, cfg.Server.IsEnabled())
}

func TestLoadDirectory(t *testing.T) {
	cfg, err := Load("../../")

	// load root pass == okay
	assert.Nil(t, err)
	assert.NotNil(t, cfg.Slack)

	// invalid directory
	cfg, err = Load("/sdsdsdds")
	assert.EqualError(t, err, "stat /sdsdsdds: no such file or directory")
	assert.Equal(t, defaultConfig, cfg)
}

func TestLoadFile(t *testing.T) {
	// not existing file
	configPath := path.Join("..", "..", "readme.sdsdsd")
	cfg, err := Load(configPath)
	assert.Contains(t, err.Error(), "stat "+configPath+": no such file or directory")
	assert.Equal(t, defaultConfig, cfg)

	// parse invalid file
	configPath = path.Join("..", "..", "readme.md")
	cfg, err = Load(configPath)
	assert.Contains(t, err.Error(), "While parsing config: yaml")
	assert.Equal(t, defaultConfig, cfg)

	// load example file == okay
	configPath = path.Join("..", "..", "config.example.yaml")
	cfg, err = Load(configPath)
	assert.Nil(t, err)
	assert.NotNil(t, cfg.Slack)
	assert.Equal(t, "info", cfg.Logger.Level)
}

func TestEnvironment(t *testing.T) {
	os.Setenv("BOT_TIMEZONE", "test/test")
	os.Setenv("BOT_SLACK_TOKEN", "myToken")

	// load example file == okay
	cfg, err := Load("../../config.example.yaml")
	assert.Nil(t, err)
	assert.Equal(t, "test/test", cfg.Timezone)
	assert.Equal(t, "myToken", cfg.Slack.Token)
	assert.Equal(t, "info", cfg.Logger.Level)

	fmt.Println(cfg.AllowedUsers)
}
