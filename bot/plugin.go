package bot

import (
	"github.com/innogames/slack-bot/v2/client"
)

// Plugin is a extended command which can be registered to the bot at compile time
type Plugin struct {
	Init func(bot *Bot, slackClient client.SlackClient) Commands
}

var pluginList []Plugin

// RegisterPlugin registers a new plugin, also in init() time
func RegisterPlugin(plugin Plugin) {
	pluginList = append(pluginList, plugin)
}

func loadPlugins(bot *Bot, slackClient client.SlackClient) Commands {
	commands := Commands{}

	for _, plugin := range pluginList {
		commands.Merge(plugin.Init(bot, slackClient))
	}

	clear(pluginList)

	return commands
}
