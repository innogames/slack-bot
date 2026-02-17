package command

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
)

type helpCommand struct {
	bot.BaseCommand
	commands       *bot.Commands
	sortedCommands []bot.Help
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
		},
		{
			Command:     "help <category>",
			Description: "displays more information about one category",
			Examples: []string{
				"help jira",
			},
		},
	}
}

// ShowAll command entries and group them by "category"
func (t *helpCommand) showAll(_ matcher.Result, message msg.Message) {
	t.once.Do(t.prebuildHelp)

	t.AddReaction("💡", message)

	var text strings.Builder

	text.WriteString("Hello <@" + message.User + ">, I’m your friendly slack-bot. You want me to show you around? :smile: \n")
	fmt.Fprintf(&text, "I currently listen to the following *%d* commands:\n", len(t.sortedCommands))

	lastCategory := bot.Category{}
	for _, commandHelp := range t.sortedCommands {
		// print new category header
		if commandHelp.Category.Name != "" && lastCategory != commandHelp.Category {
			lastCategory = commandHelp.Category
			text.WriteString(t.printCategoryHeader(commandHelp))
		}

		if commandHelp.HelpURL != "" {
			fmt.Fprintf(&text, "• <%s|%s>", commandHelp.HelpURL, commandHelp.Command)
		} else {
			fmt.Fprintf(&text, "• *%s*", commandHelp.Command)
		}
		if commandHelp.Description != "" {
			fmt.Fprintf(&text, " _(%s)_", commandHelp.Description)
		}
		text.WriteString("\n")
	}

	t.SendEphemeralMessage(message, text.String())
}

func (t *helpCommand) printCategoryHeader(commandHelp bot.Help) (text string) {
	text = "\n"
	if commandHelp.Category.HelpURL != "" {
		text += fmt.Sprintf("*<%s|%s>*", commandHelp.Category.HelpURL, commandHelp.Category.Name)
	} else {
		text += fmt.Sprintf("*%s*", commandHelp.Category.Name)
	}

	if commandHelp.Category.Description != "" {
		text += fmt.Sprintf(" (_%s_)", commandHelp.Category.Description)
	}

	text += ":\n"

	return text
}

// prints details of a specific command
func (t *helpCommand) showSingleCommand(match matcher.Result, message msg.Message) {
	// compile help only once
	t.once.Do(t.prebuildHelp)

	command := strings.TrimSpace(match.GetString("command"))

	var matchedCommand bot.Help
	for _, cmd := range t.sortedCommands {
		if strings.HasPrefix(cmd.Command, command) {
			matchedCommand = cmd
			break
		}
	}
	if matchedCommand.Command == "" {
		t.SendEphemeralMessage(message, fmt.Sprintf("Invalid command: `%s`", command))
		return
	}

	var text strings.Builder
	fmt.Fprintf(&text, "*%s*:\n", matchedCommand.Command)
	if matchedCommand.Description != "" {
		text.WriteString(matchedCommand.Description + "\n")
	}

	if len(matchedCommand.Examples) > 0 {
		text.WriteString("*Some examples:*\n")
		for _, example := range matchedCommand.Examples {
			text.WriteString(" - " + example + "\n")
		}
	}

	t.SendEphemeralMessage(message, text.String())
}

// generate the list of all commands only once and sort them by category/name
func (t *helpCommand) prebuildHelp() {
	allCommands := t.commands.GetHelp()

	sort.Slice(allCommands, func(i, j int) bool {
		if allCommands[i].Category.Name == allCommands[j].Category.Name {
			return strings.ToLower(allCommands[i].Command) < strings.ToLower(allCommands[j].Command)
		}
		return allCommands[i].Category.Name < allCommands[j].Category.Name
	})

	t.sortedCommands = allCommands
}
