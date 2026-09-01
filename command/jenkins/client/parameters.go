package client

import (
	"fmt"
	"slices"
	"strings"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/innogames/slack-bot/v2/client/vcs"
)

// Parameters is a simple string map of all build parameters
type Parameters map[string]string

func (p Parameters) String() string {
	var result strings.Builder
	for key, value := range p {
		if key == slackUserParameter || key == util.FullMatch {
			continue
		}

		result.WriteString(key + ": '" + value + "' ")
	}
	out := result.String()

	if out == "" {
		return "-none-"
	}

	return strings.TrimSpace(out)
}

// ParameterModifier are functions to mutate given Jenkins parameters
// e.g. ensure the parameter is a real "boolean" value
type ParameterModifier func(string) (string, error)

var parameterModifier = map[string]ParameterModifier{
	"branch": vcs.GetMatchingBranch,
	"lowerCase": func(input string) (string, error) {
		return strings.ToLower(input), nil
	},
	"upperCase": func(input string) (string, error) {
		return strings.ToUpper(input), nil
	},
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
// input can either be positional ("value1 value2 ...", matched against jobConfig.Parameters in order)
// or named ("NAME1=value1 NAME2=value2 ..."), which may be given in any order and interspersed
// with arbitrary connector words (e.g. "with params NAME1=value1 NAME2=value2")
func ParseParameters(jobConfig config.JobConfig, parameterString string, params Parameters) error {
	rawTokens := parseWords(parameterString)

	validNames := make(map[string]bool, len(jobConfig.Parameters))
	for _, parameterConfig := range jobConfig.Parameters {
		validNames[parameterConfig.Name] = true
	}

	namedValues := make(map[string]string)
	positional := make([]string, 0, len(rawTokens))
	hasNamedParam := false

	for _, token := range rawTokens {
		if name, value, ok := splitKeyValue(token); ok && validNames[name] {
			if !hasNamedParam {
				// tokens before the first named one (e.g. "with", "params") are just connector words
				positional = positional[:0]
			}
			namedValues[name] = value
			hasNamedParam = true
			continue
		}
		positional = append(positional, token)
	}

	if hasNamedParam {
		// in named mode, key=value tokens with unknown keys are ignored instead of used positionally
		positional = slices.DeleteFunc(positional, func(token string) bool {
			_, _, ok := splitKeyValue(token)
			return ok
		})
	}

	givenParameters := positional

	var err error
	posIndex := 0
	for _, parameterConfig := range jobConfig.Parameters {
		var value string
		if namedValue, ok := namedValues[parameterConfig.Name]; ok {
			// parameterName given as NAME=value and value can be empty
			value = namedValue
		} else if posIndex < len(givenParameters) {
			// parameterName given positionally in string
			value = givenParameters[posIndex]
			posIndex++
		} else if paramValue, ok := params[parameterConfig.Name]; ok && paramValue != "" {
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

// splitKeyValue splits a "NAME=value" token, stripping surrounding quotes from the value.
// returns ok=false if the token contains no "=".
func splitKeyValue(token string) (name, value string, ok bool) {
	idx := strings.Index(token, "=")
	if idx <= 0 {
		return "", "", false
	}

	name = token[:idx]
	value = strings.Trim(token[idx+1:], `'"`)

	return name, value, true
}

// todo cleanup, is there a nice tokenizer in place somewhere?
// 'test "foo bar" 12' -> ["test", "foo bar" "12"]
func parseWords(parameterString string) []string {
	parameters := make([]string, 0)

	cur := strings.TrimSpace(parameterString)

	var c byte
	param := make([]byte, 0)
	isQuoted := false
	var quoteChar byte

	for len(cur) > 0 {
		c, cur = cur[0], cur[1:]
		switch {
		case (c == '"' || c == '\'') && (!isQuoted || c == quoteChar):
			if isQuoted {
				isQuoted = false
				parameters = append(parameters, string(param))
				param = param[:0]
			} else {
				isQuoted = true
				quoteChar = c
			}
		case c == ' ' && !isQuoted:
			// next param
			if len(param) > 0 {
				parameters = append(parameters, string(param))
			}
			param = param[:0]
		default:
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
