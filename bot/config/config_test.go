package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestLoadExampleConfig(t *testing.T) {
	cfg, err := Load("../../config.example.yaml")
	assert.Nil(t, err)
	assert.NotNil(t, cfg.Slack)
	assert.NotEmpty(t, cfg.Macros)

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
	cfg, err := Load("../../readme.sdsdsd")
	assert.Contains(t, err.Error(), "stat ../../readme.sdsdsd: no such file or directory")
	assert.Equal(t, defaultConfig, cfg)

	// parse invalid file
	cfg, err = Load("../../readme.md")
	assert.Contains(t, err.Error(), "While parsing config: yaml")
	assert.Equal(t, defaultConfig, cfg)

	// load example file == okay
	cfg, err = Load("../../config.example.yaml")
	assert.Nil(t, err)
	assert.NotNil(t, cfg.Slack)
}

func TestEnvironment(t *testing.T) {
	os.Setenv("BOT_TIMEZONE", "test/test")
	os.Setenv("BOT_SLACK_TOKEN", "myToken")

	// load example file == okay
	cfg, err := Load("../../config.example.yaml")
	assert.Nil(t, err)
	assert.Equal(t, "test/test", cfg.Timezone)
	assert.Equal(t, "myToken", cfg.Slack.Token)

}
