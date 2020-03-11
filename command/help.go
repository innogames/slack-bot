package command

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/slack-go/slack"
	"sort"
	"strings"
	"sync"
)

// NewHelpCommand provides information about all registered commands with description and examples
func NewHelpCommand(slackClient client.SlackClient, commands *bot.Commands) bot.Command {
	return &helpCommand{slackClient: slackClient, commands: commands}
}

func (t *helpCommand) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher("help ?(?P<command>.*)", t.Run)
}

type helpCommand struct {
	slackClient  client.SlackClient
	commands     *bot.Commands
	commandNames []string
	compiledHelp map[string]bot.Help
	once         sync.Once
}

func (t *helpCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "help",
			Description: "displays all available commands",
			Examples: []string{
				"help",
				"help jira",
			},
		},
	}
}

func (t *helpCommand) Run(match matcher.Result, event slack.MessageEvent) {
	// compile help only once
	t.once.Do(t.buildHelpTree)

	command := strings.TrimSpace(match.GetString("command"))
	text := ""
	if command == "" {
		// print all command
		text = "Hello <@" + event.User + ">, I’m your friendly slack-bot. You want me to show you around?\n"
		text += "I currently listen to the following commands:\n "
		for _, name := range t.commandNames {
			text += "- *" + name + "*"
			if t.compiledHelp[name].Description != "" {
				text += " _(" + t.compiledHelp[name].Description + ")_"
			}
			text += "\n"
		}
		text += "With *help <command>_* I can provide you with more details!"
	} else {
		// print details of a specific command
		commandHelp, ok := t.compiledHelp[command]
		if !ok {
			t.slackClient.Reply(event, fmt.Sprintf("Invalid command: `%s`", command))
			return
		}

		text += fmt.Sprintf("*%s command*:\n", commandHelp.Command)
		if commandHelp.Description != "" {
			text += commandHelp.Description + "\n"
		}

		if len(commandHelp.Examples) > 0 {
			text += "*Some examples:*\n"
			for _, example := range commandHelp.Examples {
				text += " - " + example + "\n"
			}
		}
	}

	t.slackClient.Reply(event, text)
}

func (t *helpCommand) buildHelpTree() {
	var names []string
	help := map[string]bot.Help{}

	for _, commandHelp := range t.commands.GetHelp() {
		if _, ok := help[commandHelp.Command]; ok {
			// main command already defined
			continue
		}
		help[commandHelp.Command] = commandHelp
		names = append(names, commandHelp.Command)
	}

	sort.Slice(names, func(i, j int) bool {
		return strings.ToLower(names[i]) < strings.ToLower(names[j])
	})

	t.commandNames = names
	t.compiledHelp = help
}
