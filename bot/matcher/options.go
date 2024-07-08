package matcher

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/innogames/slack-bot/v2/client"

	"github.com/innogames/slack-bot/v2/bot/msg"
)

func NewOptionMatcher(baseCommand string, whiteList []string, run Runner, slackClient client.SlackClient) Matcher {
	return optionMatcher{
		command:     strings.ToLower(baseCommand),
		run:         run,
		whiteList:   whiteList,
		slackClient: slackClient,
	}
}

type optionMatcher struct {
	command     string
	run         Runner
	whiteList   []string
	slackClient client.SlackClient
}

func (m optionMatcher) Match(message msg.Message) (Runner, Result) {
	_, optionsString, ok := strings.Cut(strings.ToLower(message.Text), m.command)
	if !ok {
		// no match
		return nil, nil
	}

	options := parseOptions(optionsString)

	for option := range options {
		if !slices.Contains(m.whiteList, option) {
			return func(_ Result, message msg.Message) {
				m.slackClient.AddReaction("❌", message)
				m.slackClient.ReplyError(
					message,
					fmt.Errorf(
						"the option '%s' is not available in command %s (available: %s)",
						option,
						m.command,
						strings.Join(m.whiteList, ", "),
					),
				)
			}, Result{}
		}
	}

	return m.run, options
}

// parseOptions parses a string like "option1=value1 --option2='value 2' option3="value 3""
func parseOptions(given string) Result {
	re := regexp.MustCompile(`(--)?([\w\-]+)=('([^']*)'|"([^"]*)"|(\S+))|(\w+)`)
	matches := re.FindAllStringSubmatch(given, -1)

	options := make(Result)
	for _, match := range matches {
		if match[2] != "" {
			key := match[2]
			value := match[4] + match[5] + match[6]
			options[key] = value
		} else if match[7] != "" {
			key := match[7]
			options[key] = "true"
		}
	}

	return options
}
