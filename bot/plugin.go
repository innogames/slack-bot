package bot

import (
	"plugin"
	"text/template"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/util"
	log "github.com/sirupsen/logrus"
)

func LoadPlugins(cfg *config.Config) Commands {
	commands := Commands{}

	for _, pluginPath := range cfg.Plugins {
		log.Infof("Load plugin %s...", pluginPath)

		plug, err := plugin.Open(pluginPath)
		if err != nil {
			log.Errorf("Can't load plugin %s: %s", pluginPath, err)
			continue
		}

		templateFunctions, err := plug.Lookup("GetTemplateFunctions")
		if err == nil {
			loadTemplateFunctions(templateFunctions)
		}

		commandFunction, err := plug.Lookup("GetCommands")
		if err == nil {
			commands.Merge(loadCommands(commandFunction, cfg))
		}
	}

	return commands
}

func loadTemplateFunctions(templateFunctions plugin.Symbol) {
	functionsLookup, ok := templateFunctions.(func() template.FuncMap)
	if !ok {
		log.Error("Can't convert GetTemplateFunctions to 'func() template.FuncMap'")
		return
	}

	functions := functionsLookup()
	util.RegisterFunctions(functions)
	log.Infof("Loaded %d template functions", len(functions))
}

func loadCommands(commandFunctionsLookup plugin.Symbol, cfg *config.Config) Commands {
	commandFunctions, ok := commandFunctionsLookup.(func(*config.Config) Commands)
	if !ok {
		log.Error("Can't convert GetCommands to 'func(*config.Config) bot.Commands'")
		return Commands{}
	}

	return commandFunctions(cfg)

}
