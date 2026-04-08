package bot

import (
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/client"
	log "github.com/sirupsen/logrus"
)

// Plugin is an extended command which can be registered to the bot at compile time.
// Plugin authors create a Go module with an init() function that calls RegisterPlugin().
// Users activate plugins via blank imports: _ "github.com/company/my-plugin"
type Plugin struct {
	// Name is a unique identifier for the plugin, used in logging.
	Name string

	// Init creates the plugin's commands. It receives the SlackClient for Slack
	// interaction and the Config for reading plugin-specific configuration
	// via cfg.LoadCustom("my_plugin", &myConfig).
	Init func(slackClient client.SlackClient, cfg config.Config) Commands
}

var pluginList []Plugin

// RegisterPlugin registers a new plugin, also in init() time
func RegisterPlugin(plugin Plugin) {
	pluginList = append(pluginList, plugin)
}

func loadPlugins(slackClient client.SlackClient, cfg config.Config) Commands {
	commands := Commands{}

	for _, plugin := range pluginList {
		log.Infof("Loading plugin: %s", plugin.Name)
		commands.Merge(plugin.Init(slackClient, cfg))
	}

	pluginList = nil

	return commands
}
