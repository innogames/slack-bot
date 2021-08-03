package command

import (
	"fmt"
	"text/template"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/util"
	log "github.com/sirupsen/logrus"
)

func (c *definedCommand) GetTemplateFunction() template.FuncMap {
	customFunctions := template.FuncMap{}
	for _, function := range c.templateFunctions {
		customFunctions[function.Name] = createTemplateFunction(function)
	}

	return customFunctions
}

func createTemplateFunction(function config.TemplateFunction) func(args ...string) string {
	return func(args ...string) string {
		params := util.Parameters{}
		if len(args) < len(function.Arguments) {
			return fmt.Sprintf("%s: not enough parameters", function.Name)
		}
		for i, argName := range function.Arguments {
			params[argName] = args[i]
		}

		tmpl, err := util.CompileTemplate(function.Template)
		if err != nil {
			log.Warnf("error in template function %s: %s", function.Name, err)
			return err.Error()
		}

		response, err := util.EvalTemplate(tmpl, params)
		if err != nil {
			log.Warnf("error in template function %s: %s", function.Name, err)
			return err.Error()
		}
		return response
	}
}
