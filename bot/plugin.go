package bot

import (
	"github.com/innogames/slack-bot/v2/bot/util"
	log "github.com/sirupsen/logrus"
	"plugin"
	"text/template"
)

func loadPlugins(pluginPaths []string)  {
	for _, pluginPath := range pluginPaths {
		log.Infof("Load plugin %s...", pluginPath)

		plug, err := plugin.Open(pluginPath)
		if err != nil {
			log.Errorf("Can't load plugin %s: %s", pluginPath, err)
			continue
		}

		templateFunctions, err := plug.Lookup("GetTemplateFunctions")
		if err == nil {
			functions := templateFunctions.(func() template.FuncMap)()
			util.RegisterFunctions(functions)
			log.Infof("Loaded %d template functions", len(functions))
		}
	}
}
