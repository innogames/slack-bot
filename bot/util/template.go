package util

import (
	"bytes"
	"text/template"
	"time"
)

// TemplateFunctionProvider can be provided by Commands to register template functions to the internal parser.
// example: "{{ jiraTicketUrl $ticket.Key }}" can be used in custom commands which is provided by the "jiraCommand"
type TemplateFunctionProvider interface {
	GetTemplateFunction() template.FuncMap
}

// Parameters is a wrapper for a map[string]string which is the default set of passing template variables
type Parameters map[string]string

// some more template functions which are available in all templates
var functions = template.FuncMap{
	// creates a slice out of argument
	"makeSlice": func(args ...interface{}) []interface{} {
		return args
	},
	"slice": func(str string, start int, end int) string {
		return str[start:end]
	},
	"now": func() time.Time {
		return time.Now()
	},
}

// GetTemplateFunctions returns a list of the currently available template functions which can be used in definedCommands or user specific commands
func GetTemplateFunctions() template.FuncMap {
	return functions
}

// RegisterFunctions will add a function to any template renderer
func RegisterFunctions(funcMap template.FuncMap) {
	for name, function := range funcMap {
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
