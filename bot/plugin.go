package bot

import (
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/client"
	log "github.com/sirupsen/logrus"
)

// Plugin is an extended command which can be registered to the bot at compile time.
// Plugin authors create a Go package with an init() function that calls RegisterPlugin().
// Users activate plugins via blank imports in their main package,
// e.g. _ "github.com/innogames/slack-bot/v2/plugins/all" or _ "github.com/company/my-plugin".
// A compiled-in plugin can be disabled again with "plugins: <name>: enabled: false" in the config.
type Plugin struct {
	// Name is a unique identifier for the plugin, used in logging and as
	// config key within the "plugins:" config section.
	Name string

	// Init creates the plugin's commands. It receives the SlackClient for Slack
	// interaction and the Config for reading plugin-specific configuration
	// via cfg.LoadPlugin("my_plugin", &myConfig).
	Init func(slackClient client.SlackClient, cfg config.Config) Commands
}

var pluginList []Plugin

// RegisterPlugin registers a new plugin, also in init() time
func RegisterPlugin(plugin Plugin) {
	for _, registered := range pluginList {
		if registered.Name == plugin.Name {
			log.Warnf("Plugin %s is registered multiple times", plugin.Name)
		}
	}
	pluginList = append(pluginList, plugin)
}

// loadPlugins initializes all registered plugins which are not disabled via config.
// It consumes the plugin list: a second call within the same process loads nothing.
func loadPlugins(slackClient client.SlackClient, cfg config.Config) Commands {
	commands := Commands{}

	for _, plugin := range pluginList {
		if !cfg.IsPluginEnabled(plugin.Name) {
			log.Infof("Plugin %s is disabled via config", plugin.Name)
			continue
		}
		log.Infof("Loading plugin: %s", plugin.Name)
		commands.Merge(plugin.Init(slackClient, cfg))
	}

	pluginList = nil

	return commands
}
