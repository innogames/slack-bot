package command

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/innogames/slack-bot/client"
	log "github.com/sirupsen/logrus"
)

// NewCommands defines custom commands by defining a trigger (regexp) and a list of commands which should be executed
// it also supports placeholders by {{ .param }} using the regexp group name
func NewCommands(base bot.BaseCommand, macros []config.Command) bot.Command {
	commands := make([]command, len(macros))

	for i, macro := range macros {
		commands[i] = command{
			re:     util.CompileRegexp(macro.Trigger),
			config: macro,
		}
	}

	return &definedCommand{
		base,
		commands,
	}
}

type definedCommand struct {
	bot.BaseCommand

	// precompiled regexp and list of commands
	commands []command
}

type command struct {
	re     *regexp.Regexp
	config config.Command
}

func (c *definedCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewTextMatcher("list template functions", c.listTemplateFunction),
		matcher.WildcardMatcher(c.execute),
	)
}

func (c *definedCommand) execute(ref msg.Ref, text string) bool {
	for _, macro := range c.commands {
		match := macro.re.FindStringSubmatch(text)
		if len(match) == 0 {
			continue
		}

		// extract the parameters from regexp
		params := util.RegexpResultToParams(macro.re, match)
		params["userId"] = ref.GetUser()

		for _, commandText := range macro.config.Commands {
			command, err := util.CompileTemplate(commandText)
			if err != nil {
				log.Warnf("cannot parse command %s: %s", commandText, err.Error())
				c.ReplyError(ref, err)

				continue
			}

			text, err := util.EvalTemplate(command, params)
			if err != nil {
				log.Errorf("cannot executing command %s: %s", commandText, err.Error())
				c.ReplyError(ref, err)

				continue
			}

			// each line is interpreted as command
			for _, part := range strings.Split(text, "\n") {
				client.HandleMessage(ref.WithText(part))
			}
		}

		return true
	}

	return false
}

// receiver for "list template functions". Sort all template functions by name
func (c *definedCommand) listTemplateFunction(match matcher.Result, message msg.Message) {
	functions := util.GetTemplateFunctions()
	functionNames := make([]string, 0, len(functions))

	for name := range functions {
		functionNames = append(functionNames, name)
	}

	sort.Strings(functionNames)

	text := fmt.Sprintf("This %d are available template functions:\n", len(functions))
	for _, name := range functionNames {
		text += fmt.Sprintf("- %s\n", name)
	}

	c.SendMessage(message, text)
}

func (c *definedCommand) GetHelp() []bot.Help {
	help := make([]bot.Help, 0, len(c.commands))

	for _, macro := range c.commands {
		var category bot.Category
		if macro.config.Category != "" {
			category = bot.Category{
				Name: macro.config.Category,
			}
		}
		patternHelp := bot.Help{
			Command:     macro.config.Name,
			Description: macro.config.Description,
			Examples:    macro.config.Examples,
			Category:    category,
		}

		// as fallback use the command regexp as example
		if len(macro.config.Examples) == 0 {
			patternHelp.Examples = []string{
				macro.config.Trigger,
			}
		}
		help = append(help, patternHelp)
	}

	help = append(help, bot.Help{
		Command:     "list template functions",
		Description: "lists all available template functions for custom commands (global ones and user specific ones)",
	})

	return help
}
