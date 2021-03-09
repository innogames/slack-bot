package command

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/bot/msg"
)

type helpCommand struct {
	bot.BaseCommand
	commands       *bot.Commands
	sortedCommands []bot.Help
	commandHelp    map[string]bot.Help
	once           sync.Once
}

// NewHelpCommand provides information about all registered commands with description and examples
func NewHelpCommand(base bot.BaseCommand, commands *bot.Commands) bot.Command {
	return &helpCommand{BaseCommand: base, commands: commands}
}

func (t *helpCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewTextMatcher("help", t.showAll),
		matcher.NewRegexpMatcher("help (?P<command>.*)", t.showSingleCommand),
	)
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

// ShowAll command entries and group them by "category"
func (t *helpCommand) showAll(match matcher.Result, message msg.Message) {
	t.once.Do(t.prebuildHelp)

	var text string

	text += "Hello <@" + message.User + ">, I’m your friendly slack-bot. You want me to show you around? :smile: \n"
	text += "I currently listen to the following commands:\n"

	var lastCategory = bot.Category{}
	for _, commandHelp := range t.sortedCommands {
		// print new category header
		if commandHelp.Category.Name != "" && lastCategory != commandHelp.Category {
			lastCategory = commandHelp.Category
			text += t.printCategoryHeader(commandHelp)
		}

		if commandHelp.HelpURL != "" {
			text += fmt.Sprintf("• <%s|%s>", commandHelp.HelpURL, commandHelp.Command)
		} else {
			text += fmt.Sprintf("• *%s*", commandHelp.Command)
		}
		if commandHelp.Description != "" {
			text += fmt.Sprintf(" _(%s)_", commandHelp.Description)
		}
		text += "\n"
	}

	text += "With *help <command>* I can provide you with more details!"
	t.SendMessage(message, text)
}

func (t *helpCommand) printCategoryHeader(commandHelp bot.Help) (text string) {
	if commandHelp.Category.HelpURL != "" {
		text += fmt.Sprintf("*<%s|%s>*", commandHelp.Category.HelpURL, commandHelp.Category.Name)
	} else {
		text += fmt.Sprintf("*%s*", commandHelp.Category.Name)
	}

	if commandHelp.Category.Description != "" {
		text += fmt.Sprintf(" (_%s_)", commandHelp.Category.Description)
	}

	text += ":\n"

	return
}

// ShowSingleCommand prints details of a specific command
func (t *helpCommand) showSingleCommand(match matcher.Result, message msg.Message) {
	// compile help only once
	t.once.Do(t.prebuildHelp)

	command := strings.TrimSpace(match.GetString("command"))

	commandHelp, ok := t.commandHelp[command]
	if !ok {
		t.SendMessage(message, fmt.Sprintf("Invalid command: `%s`", command))
		return
	}

	text := fmt.Sprintf("*%s command*:\n", commandHelp.Command)
	if commandHelp.Description != "" {
		text += commandHelp.Description + "\n"
	}

	if len(commandHelp.Examples) > 0 {
		text += "*Some examples:*\n"
		for _, example := range commandHelp.Examples {
			text += " - " + example + "\n"
		}
	}

	t.SendMessage(message, text)
}

// generate the list of all commands only once and sort them by category/name
func (t *helpCommand) prebuildHelp() {
	allCommands := make([]bot.Help, 0)
	commandMap := map[string]bot.Help{}

	for _, commandHelp := range t.commands.GetHelp() {
		commandMap[commandHelp.Command] = commandHelp
		allCommands = append(allCommands, commandHelp)
	}

	sort.Slice(allCommands, func(i, j int) bool {
		if allCommands[i].Category.Name == allCommands[j].Category.Name {
			return strings.ToLower(allCommands[i].Command) < strings.ToLower(allCommands[j].Command)
		}
		return allCommands[i].Category.Name < allCommands[j].Category.Name
	})

	t.sortedCommands = allCommands
	t.commandHelp = commandMap
}
