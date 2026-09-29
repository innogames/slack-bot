package bot

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/client"
	log "github.com/sirupsen/logrus"
)

// Plugin is an extension of the bot which is compiled into the bot binary, usually via the slack-bot-builder.
// Plugin authors create a Go package with an init() function which calls RegisterPlugin().
// A compiled-in plugin is only loaded when it's listed in the "plugins" section of the config, see docs/plugins.md
type Plugin struct {
	// Name is the unique identifier of the plugin, it's also the key within the "plugins" config section
	Name string

	// Setup creates the commands of the plugin. It's only called when the plugin is enabled in the config.
	Setup func(ctx *PluginContext) (Commands, error)

	// Init creates the commands of the plugin.
	//
	// Deprecated: use Setup, which also provides the plugin specific config via PluginContext.LoadConfig
	Init func(slackClient client.SlackClient, cfg config.Config) Commands
}

// PluginContext provides access to the bot framework for a plugin. Other parts of the framework, like the storage
// ("bot/storage"), the queue ("command/queue") or stats ("bot/stats"), are packages which can be used directly.
type PluginContext struct {
	// Name of the plugin
	Name string

	// SlackClient is the common interface to interact with Slack, it's also embedded in BaseCommand
	SlackClient client.SlackClient

	// Slack is the full Slack client, providing the whole Slack API via the embedded *slack.Client and the socket mode client.
	// It's nil when the SlackClient is mocked, e.g. in unit tests.
	Slack *client.Slack

	// Config is the whole bot config
	Config config.Config

	// Logger which adds the plugin name to all log entries
	Logger *log.Entry

	pluginConfig config.PluginConfig
}

// NewPluginContext creates the context for the given plugin. It's used to load the plugins and can be used to test them.
func NewPluginContext(name string, slackClient client.SlackClient, cfg config.Config) *PluginContext {
	slackAPI, _ := slackClient.(*client.Slack)

	return &PluginContext{
		Name:         name,
		SlackClient:  slackClient,
		Slack:        slackAPI,
		Config:       cfg,
		Logger:       log.WithField("plugin", name),
		pluginConfig: cfg.Plugins[name],
	}
}

// LoadConfig decodes the plugin specific options ("config" of the plugin in the "plugins" config section) into the target.
// "${ENV_VAR}" placeholders get replaced by the value of the environment variable.
func (c *PluginContext) LoadConfig(target any) error {
	return c.pluginConfig.Decode(target)
}

// BaseCommand returns a BaseCommand which can be embedded into the commands of the plugin
func (c *PluginContext) BaseCommand() BaseCommand {
	return BaseCommand{SlackClient: c.SlackClient}
}

// legacyConfigKeys maps outdated top-level config keys to the plugins which are providing these commands now
var legacyConfigKeys = map[string]string{
	"aws":          "aws",
	"ripeatlas":    "ripeatlas",
	"open_weather": "weather",
}

var pluginList []Plugin

// RegisterPlugin registers a new plugin, usually called in the init() function of the plugin package
func RegisterPlugin(plugin Plugin) {
	if plugin.Name == "" {
		log.Warn("A plugin without a name can't be registered")
		return
	}

	for _, registered := range pluginList {
		if registered.Name == plugin.Name {
			log.Warnf("Plugin %s is registered multiple times", plugin.Name)
		}
	}

	pluginList = append(pluginList, plugin)
}

// loadPlugins initializes all compiled-in plugins which are enabled in the "plugins" config section
func loadPlugins(slackClient client.SlackClient, cfg config.Config) Commands {
	commands := Commands{}
	registered := make(map[string]bool, len(pluginList))

	for _, plugin := range pluginList {
		registered[plugin.Name] = true

		pluginCfg, ok := cfg.Plugins[plugin.Name]
		if !ok {
			log.Infof("Plugin %s is compiled in, but not enabled in the \"plugins\" config section", plugin.Name)
			continue
		}
		if !pluginCfg.IsEnabled() {
			log.Infof("Plugin %s is disabled via config", plugin.Name)
			continue
		}

		pluginCommands, err := setupPlugin(plugin, NewPluginContext(plugin.Name, slackClient, cfg))
		if err != nil {
			log.Errorf("Failed to load plugin %s: %s", plugin.Name, err)
			continue
		}

		log.Infof("Loaded plugin %s with %d commands", plugin.Name, pluginCommands.Count())
		commands.Merge(pluginCommands)
	}

	for _, name := range slices.Sorted(maps.Keys(cfg.Plugins)) {
		if !registered[name] {
			log.Errorf(
				"Plugin %s is configured, but not compiled into this bot binary (compiled-in plugins: %s). "+
					"Build a bot binary including this plugin with the slack-bot-builder, see docs/plugins.md",
				name,
				formatPluginNames(slices.Sorted(maps.Keys(registered))),
			)
		}
	}

	for _, key := range slices.Sorted(maps.Keys(legacyConfigKeys)) {
		if cfg.IsSet(key) {
			log.Warnf(
				"The %q config section is not used anymore: the commands moved into the %q plugin. "+
					"Add the plugin to the \"plugins\" config section and move the options to \"plugins.%s.config\", see docs/plugins.md",
				key,
				legacyConfigKeys[key],
				legacyConfigKeys[key],
			)
		}
	}

	return commands
}

// setupPlugin creates the commands of the plugin, a panic within the plugin is returned as error
func setupPlugin(plugin Plugin, ctx *PluginContext) (commands Commands, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in plugin setup: %v", r)
		}
	}()

	switch {
	case plugin.Setup != nil:
		return plugin.Setup(ctx)
	case plugin.Init != nil:
		return plugin.Init(ctx.SlackClient, ctx.Config), nil
	default:
		return commands, errors.New("plugin has no Setup function")
	}
}

func formatPluginNames(names []string) string {
	if len(names) == 0 {
		return "none"
	}

	return strings.Join(names, ", ")
}
