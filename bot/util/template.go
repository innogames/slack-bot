package util

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
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

var functions = template.FuncMap{
	// creates a slice out of argument
	"makeSlice": func(args ...any) []any {
		return args
	},
	"makeMap": func(args ...any) (map[string]any, error) {
		if len(args)%2 != 0 {
			return nil, errors.New("makeMap: expected alternating key-value pairs as arguments")
		}

		m := make(map[string]any, len(args)/2)
		for i := 0; i < len(args); i += 2 {
			key, ok := args[i].(string)
			if !ok {
				return nil, fmt.Errorf("makeMap: arg at index %d: key must be string", i)
			}

			val := args[i+1]

			m[key] = val
		}

		return m, nil
	},
	"slice": func(str string, start int, end int) string {
		return str[start:end]
	},
	"date": func(date string, inFormat string, outFormat string) (string, error) {
		t, err := time.Parse(inFormat, date)
		if err != nil {
			return "invalid format", err
		}
		return t.In(time.Local).Format(outFormat), nil
	},
}

// GetTemplateFunctions returns a list of the currently available template functions which can be used in definedCommands or user specific commands
func GetTemplateFunctions() template.FuncMap {
	return functions
}

// RegisterFunctions will add a function to any template renderer
func RegisterFunctions(funcMap template.FuncMap) {
	maps.Copy(functions, funcMap)
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
