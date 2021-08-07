package plugins

import (
	"plugin"
	"text/template"

	"github.com/innogames/slack-bot/v2/bot/util"
	log "github.com/sirupsen/logrus"
)

func LoadPlugins(pluginPaths []string) {
	for _, pluginPath := range pluginPaths {
		loadPlugin(pluginPath)
	}
}

func loadPlugin(pluginPath string) {
	log.Infof("Load plugin %s...", pluginPath)

	plug, err := plugin.Open(pluginPath)
	if err != nil {
		log.Errorf("Can't load plugin %s: %s", pluginPath, err)
		return
	}

	templateFunctions, err := plug.Lookup("GetTemplateFunctions")
	if err == nil {
		functionsLookup, ok := templateFunctions.(func() template.FuncMap)
		if !ok {
			log.Error("Can't convert GetTemplateFunctions to template.FuncMap")
			return
		}

		functions := functionsLookup()
		util.RegisterFunctions(functions)
		log.Infof("Loaded %d template functions", len(functions))
	}
}
