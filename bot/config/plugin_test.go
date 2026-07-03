package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testPluginConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	APIKey  string `mapstructure:"api_key"`
}

func TestLoadPlugin(t *testing.T) {
	t.Run("nil viper", func(t *testing.T) {
		cfg := Config{}

		pluginCfg := testPluginConfig{APIKey: "default"}
		require.NoError(t, cfg.LoadPlugin("my_plugin", &pluginCfg))

		// defaults are untouched
		assert.Equal(t, "default", pluginCfg.APIKey)
	})

	t.Run("load from plugins section", func(t *testing.T) {
		cfg := Config{}
		cfg.Set("plugins§my_plugin", map[string]any{
			"enabled": true,
			"api_key": "secret",
		})

		pluginCfg := testPluginConfig{}
		require.NoError(t, cfg.LoadPlugin("my_plugin", &pluginCfg))

		assert.True(t, pluginCfg.Enabled)
		assert.Equal(t, "secret", pluginCfg.APIKey)
	})

	t.Run("fallback to legacy top-level key", func(t *testing.T) {
		cfg := Config{}
		cfg.Set("my_plugin", map[string]any{
			"api_key": "legacy-secret",
		})

		pluginCfg := testPluginConfig{}
		require.NoError(t, cfg.LoadPlugin("my_plugin", &pluginCfg))

		assert.Equal(t, "legacy-secret", pluginCfg.APIKey)
	})

	t.Run("plugins section wins over legacy key", func(t *testing.T) {
		cfg := Config{}
		cfg.Set("my_plugin", map[string]any{
			"api_key": "legacy-secret",
		})
		cfg.Set("plugins§my_plugin", map[string]any{
			"api_key": "new-secret",
		})

		pluginCfg := testPluginConfig{}
		require.NoError(t, cfg.LoadPlugin("my_plugin", &pluginCfg))

		assert.Equal(t, "new-secret", pluginCfg.APIKey)
	})

	t.Run("absent config keeps defaults", func(t *testing.T) {
		cfg := Config{}
		cfg.Set("other_plugin", map[string]any{"api_key": "other"})

		pluginCfg := testPluginConfig{APIKey: "default"}
		require.NoError(t, cfg.LoadPlugin("my_plugin", &pluginCfg))

		assert.Equal(t, "default", pluginCfg.APIKey)
	})
}

func TestIsPluginEnabled(t *testing.T) {
	t.Run("nil viper", func(t *testing.T) {
		cfg := Config{}
		assert.True(t, cfg.IsPluginEnabled("my_plugin"))
	})

	t.Run("no plugins section", func(t *testing.T) {
		cfg := Config{}
		cfg.Set("something", "else")
		assert.True(t, cfg.IsPluginEnabled("my_plugin"))
	})

	t.Run("section without enabled flag", func(t *testing.T) {
		cfg := Config{}
		cfg.Set("plugins§my_plugin", map[string]any{"api_key": "secret"})
		assert.True(t, cfg.IsPluginEnabled("my_plugin"))
	})

	t.Run("explicitly disabled", func(t *testing.T) {
		cfg := Config{}
		cfg.Set("plugins§my_plugin", map[string]any{"enabled": false})
		assert.False(t, cfg.IsPluginEnabled("my_plugin"))
	})

	t.Run("explicitly enabled", func(t *testing.T) {
		cfg := Config{}
		cfg.Set("plugins§my_plugin", map[string]any{"enabled": true})
		assert.True(t, cfg.IsPluginEnabled("my_plugin"))
	})
}

// the whole plugin config design relies on nested key traversal (IsSet/UnmarshalKey
// with the "§" delimiter) working in the viper fork - lock that behavior in
func TestNestedViperKeys(t *testing.T) {
	cfg := Config{}
	cfg.Set("plugins§my_plugin", map[string]any{"enabled": false, "api_key": "secret"})

	assert.True(t, cfg.viper.IsSet("plugins§my_plugin"))
	assert.True(t, cfg.viper.IsSet("plugins§my_plugin§enabled"))
	assert.False(t, cfg.viper.IsSet("plugins§other_plugin"))
	assert.False(t, cfg.viper.GetBool("plugins§my_plugin§enabled"))
	assert.Equal(t, "secret", cfg.viper.GetString("plugins§my_plugin§api_key"))
}
