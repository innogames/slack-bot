package command

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/innogames/slack-bot/client"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"regexp"
	"strings"
)

// NewMacroCommand defines custom commands by defining a trigger (regexp) and a list of commands which should be executed
// it also supports placeholders by {{ .param }} using the regexp group name
func NewMacroCommand(slackClient client.SlackClient, macros []config.Macro, logger *logrus.Logger) bot.Command {
	commands := make([]command, len(macros))

	for i, macro := range macros {
		commands[i] = command{
			re:     util.CompileRegexp(macro.Trigger),
			config: macro,
		}
	}

	return &macroCommand{
		slackClient,
		commands,
	}
}

type macroCommand struct {
	slackClient client.SlackClient

	// precompiled regexp and list of commands
	commands []command
}

type command struct {
	re     *regexp.Regexp
	config config.Macro
}

func (c *macroCommand) GetMatcher() matcher.Matcher {
	return matcher.WildcardMatcher(c.Execute)
}

func (c *macroCommand) Execute(event slack.MessageEvent) bool {
	for _, macro := range c.commands {
		match := macro.re.FindStringSubmatch(event.Text)
		if len(match) == 0 {
			continue
		}

		// extract the parameters from regexp
		params := util.RegexpResultToParams(macro.re, match)
		params["userId"] = event.User

		for _, commandText := range macro.config.Commands {
			command, err := util.CompileTemplate(commandText)
			if err != nil {
				fmt.Printf("cannot parse command %s: %s\n", commandText, err.Error())
				c.slackClient.ReplyError(event, err)
				continue
			}

			text, err := util.EvalTemplate(command, params)
			if err != nil {
				fmt.Printf("cannot executing command %s: %s\n", commandText, err.Error())
				c.slackClient.ReplyError(event, err)
				continue
			}

			// each line is interpreted as command
			for _, part := range strings.Split(text, "\n") {
				newMessage := event
				newMessage.Text = part
				client.InternalMessages <- msg.FromSlackEvent(newMessage)
			}
		}

		return true
	}

	return false
}

func (c *macroCommand) GetHelp() []bot.Help {
	help := make([]bot.Help, 0, len(c.commands))

	for _, macro := range c.commands {
		patternHelp := bot.Help{
			Command:     macro.config.Name,
			Description: macro.config.Description,
			Examples:    macro.config.Examples,
		}

		// as fallback use the command regexp as example
		if len(macro.config.Examples) == 0 {
			patternHelp.Examples = []string{
				macro.config.Trigger,
			}
		}
		help = append(help, patternHelp)
	}

	return help
}
