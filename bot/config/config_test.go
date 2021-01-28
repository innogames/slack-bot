package config

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadExampleConfig(t *testing.T) {
	cfg, err := Load("../../config.example.yaml")
	assert.Nil(t, err)
	assert.NotNil(t, cfg.Slack)
	assert.NotEmpty(t, cfg.Commands)

	assert.Equal(t, false, cfg.Jenkins.IsEnabled())
	assert.Equal(t, false, cfg.Jira.IsEnabled())
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
	assert.NotNil(t, err)
	assert.Equal(t, defaultConfig, cfg)
}

func TestLoadFile(t *testing.T) {
	// not existing file
	configPath := path.Join("..", "..", "readme.sdsdsd")
	cfg, err := Load(configPath)
	assert.NotNil(t, err)
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
	os.Setenv("BOT_TIMEZONE", "Europe/Berlin")
	os.Setenv("BOT_SLACK_TOKEN", "mySlackToken")
	os.Setenv("BOT_GITHUB_ACCESS_TOKEN", "myGithubToken")

	// load example file == okay
	cfg, err := Load("../../config.example.yaml")
	assert.Nil(t, err)
	assert.Equal(t, "Europe/Berlin", cfg.Timezone)
	assert.Equal(t, "mySlackToken", cfg.Slack.Token)
	assert.Equal(t, "info", cfg.Logger.Level)
	assert.Equal(t, "myGithubToken", cfg.Github.AccessToken)
}
