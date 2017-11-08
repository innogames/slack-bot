package util

import (
	"bytes"
	"text/template"
)

type TemplateFunctionProvider interface {
	GetTemplateFunction() template.FuncMap
}

type Parameters map[string]string

var functions = template.FuncMap{
	// creates a slice out of argument
	"makeSlice": func(args ...interface{}) []interface{} {
		return args
	},
	"slice": func(string string, start int, end int) string {
		return string[start:end]
	},
}

// RegisterFunctions will add a function to any template renderer
func RegisterFunctions(provider TemplateFunctionProvider) {
	for name, function := range provider.GetTemplateFunction() {
		functions[name] = function
	}
}

// CompileTemplate pre compiles a template and returns an error if an function is not available etc
func CompileTemplate(temp string) (*template.Template, error) {
	return template.New(temp).Funcs(functions).Parse(temp)
}

// EvalTemplate renders the template
func EvalTemplate(temp *template.Template, params Parameters) (string, error) {
	var buf bytes.Buffer
	err := temp.Execute(&buf, params)

	return buf.String(), err
}
