package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadPlugins(t *testing.T) {
	dir := t.TempDir()
	writeConfigFile(t, dir, "a.yaml", `
plugins:
  aws:
    source: github.com/innogames/slack-bot/plugins/aws
    version: v1.2.0
    config:
      enabled: true
      cloud_front:
        - id: E1234ABCDEF
          name: my-distribution
  deploy_tools:
    source: https://gitlab.example.com/team/slack-bot-plugins.git
    version: main
    subdir: deploy
    config:
      api_url: https://deploy.example.com
      api_token: ""
  weather:
    source: ./plugins/weather
    enabled: false
`)
	// e.g. a secrets.yaml which is not part of the config repository
	writeConfigFile(t, dir, "b.yaml", `
plugins:
  deploy_tools:
    config:
      api_token: "${DEPLOY_TOKEN}"
`)

	t.Setenv("DEPLOY_TOKEN", "secret")
	t.Setenv("BOT_PLUGINS_DEPLOY_TOOLS_CONFIG_API_URL", "https://deploy2.example.com")

	cfg, err := Load(dir)
	require.NoError(t, err)

	require.Len(t, cfg.Plugins, 3)

	aws := cfg.Plugins["aws"]
	assert.Equal(t, "github.com/innogames/slack-bot/plugins/aws", aws.Source)
	assert.Equal(t, "v1.2.0", aws.Version)
	assert.True(t, aws.IsEnabled())

	deployTools := cfg.Plugins["deploy_tools"]
	assert.Equal(t, "https://gitlab.example.com/team/slack-bot-plugins.git", deployTools.Source)
	assert.Equal(t, "main", deployTools.Version)
	assert.Equal(t, "deploy", deployTools.Subdir)
	assert.True(t, deployTools.IsEnabled())

	deployConfig := struct {
		APIURL   string `mapstructure:"api_url"`
		APIToken string `mapstructure:"api_token"`
	}{}
	require.NoError(t, deployTools.Decode(&deployConfig))
	assert.Equal(t, "https://deploy2.example.com", deployConfig.APIURL)
	assert.Equal(t, "secret", deployConfig.APIToken)

	// the raw config still contains the placeholder
	assert.Equal(t, "${DEPLOY_TOKEN}", deployTools.Config["api_token"])

	weather := cfg.Plugins["weather"]
	assert.Equal(t, "./plugins/weather", weather.Source)
	assert.False(t, weather.IsEnabled())
	require.NoError(t, weather.Decode(&deployConfig))
}

func TestExpandEnv(t *testing.T) {
	t.Setenv("PLUGIN_USER", "bot")
	t.Setenv("PLUGIN_PASSWORD", "pa$$word")

	input := map[string]any{
		"user":     "${PLUGIN_USER}",
		"password": "${PLUGIN_PASSWORD}",
		"url":      "https://${PLUGIN_USER}@example.com/$path/${UNKNOWN_PLUGIN_VAR}",
		"nested": map[string]any{
			"list": []any{"${PLUGIN_USER}", 12, true},
		},
		"number": 42,
	}

	expected := map[string]any{
		"user":     "bot",
		"password": "pa$$word",
		"url":      "https://bot@example.com/$path/",
		"nested": map[string]any{
			"list": []any{"bot", 12, true},
		},
		"number": 42,
	}

	assert.Equal(t, expected, expandEnv(input))

	// the input is not modified
	assert.Equal(t, "${PLUGIN_USER}", input["user"])
	assert.Equal(t, "${PLUGIN_USER}", input["nested"].(map[string]any)["list"].([]any)[0])
}

func TestIsSet(t *testing.T) {
	cfg := Config{}
	assert.False(t, cfg.IsSet("aws"))

	cfg.Set("aws.enabled", false)
	assert.True(t, cfg.IsSet("aws"))
	assert.True(t, cfg.IsSet("aws.enabled"))
	assert.False(t, cfg.IsSet("aws.cloud_front"))
	assert.False(t, cfg.IsSet("aws.enabled.foo"))
}
