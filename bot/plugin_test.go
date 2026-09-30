package bot

import (
	"errors"
	"testing"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadPlugins(t *testing.T) {
	slackClient := mocks.NewSlackClient(t)

	originalPlugins := pluginList
	t.Cleanup(func() {
		pluginList = originalPlugins
	})

	logHook := test.NewGlobal()
	t.Cleanup(func() {
		logrus.StandardLogger().ReplaceHooks(logrus.LevelHooks{})
	})

	// registers a single plugin which counts its Setup calls
	registerPlugin := func(name string) *int {
		setupCalls := 0

		pluginList = nil
		RegisterPlugin(Plugin{
			Name: name,
			Setup: func(ctx *PluginContext) (Commands, error) {
				setupCalls++
				assert.Equal(t, name, ctx.Name)
				assert.Equal(t, slackClient, ctx.SlackClient)
				assert.Nil(t, ctx.Slack)

				commands := Commands{}
				commands.AddCommand(testCommand2{})

				return commands, nil
			},
		})

		return &setupCalls
	}

	enabled := func(enabled bool) *bool {
		return &enabled
	}

	t.Run("plugin is not loaded without config", func(t *testing.T) {
		setupCalls := registerPlugin("fake")

		commands := loadPlugins(slackClient, config.Config{})

		assert.Equal(t, 0, *setupCalls)
		assert.Equal(t, 0, commands.Count())
	})

	t.Run("plugin is loaded when configured", func(t *testing.T) {
		setupCalls := registerPlugin("fake")

		cfg := config.Config{
			Plugins: map[string]config.PluginConfig{
				"fake": {},
			},
		}
		commands := loadPlugins(slackClient, cfg)

		assert.Equal(t, 1, *setupCalls)
		assert.Equal(t, 1, commands.Count())

		// the registry is kept, e.g. for a second bot instance
		commands = loadPlugins(slackClient, cfg)
		assert.Equal(t, 2, *setupCalls)
		assert.Equal(t, 1, commands.Count())
	})

	t.Run("plugin is loaded when enabled explicitly", func(t *testing.T) {
		setupCalls := registerPlugin("fake")

		cfg := config.Config{
			Plugins: map[string]config.PluginConfig{
				"fake": {Enabled: enabled(true)},
			},
		}
		commands := loadPlugins(slackClient, cfg)

		assert.Equal(t, 1, *setupCalls)
		assert.Equal(t, 1, commands.Count())
	})

	t.Run("plugin is not loaded when disabled", func(t *testing.T) {
		setupCalls := registerPlugin("fake")

		cfg := config.Config{
			Plugins: map[string]config.PluginConfig{
				"fake": {Enabled: enabled(false)},
			},
		}
		commands := loadPlugins(slackClient, cfg)

		assert.Equal(t, 0, *setupCalls)
		assert.Equal(t, 0, commands.Count())
	})

	t.Run("configured plugin is not compiled in", func(t *testing.T) {
		registerPlugin("fake")
		logHook.Reset()

		cfg := config.Config{
			Plugins: map[string]config.PluginConfig{
				"unknown": {},
			},
		}
		commands := loadPlugins(slackClient, cfg)

		assert.Equal(t, 0, commands.Count())
		require.NotNil(t, logHook.LastEntry())
		assert.Equal(t, logrus.ErrorLevel, logHook.LastEntry().Level)
		assert.Contains(t, logHook.LastEntry().Message, "Plugin unknown is configured, but not compiled into this bot binary (compiled-in plugins: fake)")
	})

	t.Run("deprecated Init function", func(t *testing.T) {
		initCalls := 0

		pluginList = nil
		RegisterPlugin(Plugin{
			Name: "legacy",
			Init: func(_ client.SlackClient, _ config.Config) Commands {
				initCalls++

				commands := Commands{}
				commands.AddCommand(testCommand2{})

				return commands
			},
		})

		cfg := config.Config{
			Plugins: map[string]config.PluginConfig{
				"legacy": {},
			},
		}
		commands := loadPlugins(slackClient, cfg)

		assert.Equal(t, 1, initCalls)
		assert.Equal(t, 1, commands.Count())
	})

	t.Run("failing plugins are skipped", func(t *testing.T) {
		pluginList = nil
		RegisterPlugin(Plugin{
			Name: "error",
			Setup: func(_ *PluginContext) (Commands, error) {
				return Commands{}, errors.New("invalid api key")
			},
		})
		RegisterPlugin(Plugin{
			Name: "panic",
			Setup: func(_ *PluginContext) (Commands, error) {
				panic("oops")
			},
		})
		RegisterPlugin(Plugin{
			Name: "empty",
		})
		logHook.Reset()

		cfg := config.Config{
			Plugins: map[string]config.PluginConfig{
				"error": {},
				"panic": {},
				"empty": {},
			},
		}
		commands := loadPlugins(slackClient, cfg)

		assert.Equal(t, 0, commands.Count())

		entries := logHook.AllEntries()
		messages := make([]string, 0, len(entries))
		for _, entry := range entries {
			messages = append(messages, entry.Message)
		}
		assert.Equal(t, []string{
			"Failed to load plugin error: invalid api key",
			"Failed to load plugin panic: panic in plugin setup: oops",
			"Failed to load plugin empty: plugin has no Setup function",
		}, messages)
	})

	t.Run("warning for outdated config sections", func(t *testing.T) {
		pluginList = nil
		logHook.Reset()

		cfg := config.Config{}
		cfg.Set("aws.enabled", true)
		loadPlugins(slackClient, cfg)

		require.NotNil(t, logHook.LastEntry())
		assert.Equal(t, logrus.WarnLevel, logHook.LastEntry().Level)
		assert.Contains(t, logHook.LastEntry().Message, `The "aws" config section is not used anymore: the commands moved into the "aws" plugin`)
	})

	t.Run("register plugins", func(t *testing.T) {
		pluginList = nil

		RegisterPlugin(Plugin{Name: ""})
		assert.Empty(t, pluginList)

		RegisterPlugin(Plugin{Name: "dup"})
		RegisterPlugin(Plugin{Name: "dup"})
		assert.Len(t, pluginList, 2)
	})
}

func TestPluginContext(t *testing.T) {
	slackClient := mocks.NewSlackClient(t)

	type pluginConfig struct {
		Account string   `mapstructure:"account"`
		Hosts   []string `mapstructure:"hosts"`
	}

	t.Run("load plugin config", func(t *testing.T) {
		t.Setenv("PLUGIN_TEST_ACCOUNT", "bot")

		cfg := config.Config{
			Plugins: map[string]config.PluginConfig{
				"my_plugin": {
					Config: map[string]any{
						"account": "${PLUGIN_TEST_ACCOUNT}",
						"hosts":   []any{"a.example.com", "b.example.com"},
					},
				},
			},
		}

		ctx := NewPluginContext("my_plugin", slackClient, cfg)
		assert.Equal(t, "my_plugin", ctx.Name)
		assert.Equal(t, BaseCommand{SlackClient: slackClient}, ctx.BaseCommand())
		assert.Equal(t, "my_plugin", ctx.Logger.Data["plugin"])

		loaded := pluginConfig{}
		require.NoError(t, ctx.LoadConfig(&loaded))
		assert.Equal(t, pluginConfig{Account: "bot", Hosts: []string{"a.example.com", "b.example.com"}}, loaded)
	})

	t.Run("keep default values without config", func(t *testing.T) {
		ctx := NewPluginContext("my_plugin", slackClient, config.Config{})

		loaded := pluginConfig{Account: "default"}
		require.NoError(t, ctx.LoadConfig(&loaded))
		assert.Equal(t, pluginConfig{Account: "default"}, loaded)
	})

	t.Run("full slack client", func(t *testing.T) {
		slack, err := client.GetSlackClient(config.Slack{Token: "xoxb-12345", SocketToken: "xapp-6789"})
		require.NoError(t, err)

		ctx := NewPluginContext("my_plugin", slack, config.Config{})
		assert.Same(t, slack, ctx.Slack)
	})
}
