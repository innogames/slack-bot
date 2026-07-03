package bot

import (
	"testing"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
)

func TestLoadPlugins(t *testing.T) {
	slackClient := mocks.NewSlackClient(t)

	// loadPlugins consumes the global plugin list, restore whatever was registered before
	originalPlugins := pluginList
	t.Cleanup(func() {
		pluginList = originalPlugins
	})

	registerFakePlugin := func(initCalled *bool) {
		pluginList = nil
		RegisterPlugin(Plugin{
			Name: "fake",
			Init: func(_ client.SlackClient, _ config.Config) Commands {
				*initCalled = true

				commands := Commands{}
				commands.AddCommand(&testCommand2{})
				return commands
			},
		})
	}

	t.Run("plugin loads without any config", func(t *testing.T) {
		initCalled := false
		registerFakePlugin(&initCalled)

		commands := loadPlugins(slackClient, config.Config{})

		assert.True(t, initCalled)
		assert.Equal(t, 1, commands.Count())
		assert.Nil(t, pluginList)
	})

	t.Run("plugin loads when explicitly enabled", func(t *testing.T) {
		initCalled := false
		registerFakePlugin(&initCalled)

		cfg := config.Config{}
		cfg.Set("plugins§fake§enabled", true)
		commands := loadPlugins(slackClient, cfg)

		assert.True(t, initCalled)
		assert.Equal(t, 1, commands.Count())
	})

	t.Run("plugin is skipped when disabled via config", func(t *testing.T) {
		initCalled := false
		registerFakePlugin(&initCalled)

		cfg := config.Config{}
		cfg.Set("plugins§fake§enabled", false)
		commands := loadPlugins(slackClient, cfg)

		assert.False(t, initCalled)
		assert.Equal(t, 0, commands.Count())
	})

	t.Run("duplicate registration logs but keeps both", func(t *testing.T) {
		pluginList = nil
		noopInit := func(_ client.SlackClient, _ config.Config) Commands {
			return Commands{}
		}
		RegisterPlugin(Plugin{Name: "dup", Init: noopInit})
		RegisterPlugin(Plugin{Name: "dup", Init: noopInit})

		assert.Len(t, pluginList, 2)
	})
}
