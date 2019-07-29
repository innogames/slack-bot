package jenkins

import (
	"fmt"
	"github.com/innogames/slack-bot/client/vcs"
	"github.com/innogames/slack-bot/config"
	"strings"
)

// Parameters is a simple string map of all build parameters
type Parameters map[string]string

// ParameterModifier are functions to mutate given Jenkins parameters
// e.g. ensure the parameter is a real "boolean" value
type ParameterModifier func(string) (string, error)

var parameterModifier = map[string]ParameterModifier{
	"branch": vcs.GetMatchingBranch,
	"bool": func(value string) (string, error) {
		switch value {
		case "false", "FALSE", "0", "null", "", " ":
			return "false", nil
		default:
			return "true", nil
		}
	},
}

// if a job was triggered via bot we send this additional build param to jenkins with the slack user name
const slackUserParameter = "SLACK_USER"

// ParseParameters parse jenkins parameters, based on a input string
func ParseParameters(jobConfig config.JobConfig, parameterString string, params Parameters) error {
	givenParameters := parseWords(parameterString)

	var err error
	for index, parameterConfig := range jobConfig.Parameters {
		var value string
		if len(givenParameters) > index {
			// parameterName given in string
			value = givenParameters[index]
		} else if _, ok := params[parameterConfig.Name]; ok {
			// use given names parameterName!
			value = params[parameterConfig.Name]
		} else if parameterConfig.Default != "" {
			// use default value
			value = parameterConfig.Default
		} else {
			err := fmt.Errorf("sorry, you have to pass %d parameters (%s)", len(jobConfig.Parameters), strings.Join(getNames(jobConfig.Parameters), ", "))

			return err
		}

		if modifier, ok := parameterModifier[parameterConfig.Type]; ok {
			value, err = modifier(value)
			if err != nil {
				return err
			}
		}

		params[parameterConfig.Name] = value
	}

	return nil
}

// todo cleanup, is there a nice tokenizer in place somewhere?
// 'test "foo bar" 12' -> ["test", "foo bar" "12"]
func parseWords(parameterString string) []string {
	parameters := make([]string, 0)

	cur := strings.TrimSpace(parameterString)

	var c byte
	var param = make([]byte, 0)
	isQuoted := false
	for len(cur) > 0 {
		c, cur = cur[0], cur[1:]
		if c == '"' {
			if isQuoted {
				isQuoted = false
				parameters = append(parameters, string(param))
				param = make([]byte, 0)
			} else {
				isQuoted = true
			}
		} else if c == ' ' && !isQuoted {
			// next param
			parameters = append(parameters, string(param))
			param = make([]byte, 0)
		} else {
			// append char to current param
			param = append(param, c)
		}
	}

	if len(param) > 0 {
		// open quoting...just add it as last parameter
		parameters = append(parameters, string(param))
	}

	return parameters
}

func getNames(list []config.JobParameter) []string {
	keys := make([]string, len(list))

	for i, parameter := range list {
		keys[i] = parameter.Name
	}

	return keys
}
