package command

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/client"
	"io"
	"sort"
	"strings"
	"sync"
)

// NewHelpCommand provides information about all registered commands with description and examples
func NewHelpCommand(slackClient client.SlackClient, commands *bot.Commands) bot.Command {
	return &helpCommand{slackClient: slackClient, commands: commands}
}

func (t *helpCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewTextMatcher("help", t.ShowAll),
		matcher.NewRegexpMatcher("help (?P<command>.*)", t.ShowSingleCommand),
	)
}

type helpCommand struct {
	slackClient    client.SlackClient
	commands       *bot.Commands
	sortedCommands []bot.Help
	commandHelp    map[string]bot.Help
	once           sync.Once
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
func (t *helpCommand) ShowAll(match matcher.Result, message msg.Message) {
	t.once.Do(t.prebuildHelp)

	var text strings.Builder

	text.WriteString("Hello <@" + message.User + ">, I’m your friendly slack-bot. You want me to show you around? :smile: \n")
	text.WriteString("I currently listen to the following commands:\n")

	var lastCategory = bot.Category{}
	for _, commandHelp := range t.sortedCommands {
		// print new category header
		if commandHelp.Category.Name != "" && lastCategory != commandHelp.Category {
			lastCategory = commandHelp.Category
			t.printCategoryHeader(commandHelp, &text)
		}

		if commandHelp.HelpURL != "" {
			text.WriteString(fmt.Sprintf(" - <%s|%s>", commandHelp.HelpURL, commandHelp.Command))
		} else {
			text.WriteString(fmt.Sprintf("- *%s*", commandHelp.Command))
		}
		if commandHelp.Description != "" {
			text.WriteString(fmt.Sprintf(" _(%s)_", commandHelp.Description))
		}
		text.WriteString("\n")
	}

	text.WriteString("With *help <command>* I can provide you with more details!")
	t.slackClient.SendMessage(message, text.String())
}

func (t *helpCommand) printCategoryHeader(commandHelp bot.Help, text io.StringWriter) {
	if commandHelp.Category.HelpURL != "" {
		text.WriteString(fmt.Sprintf("*<%s|%s>*", commandHelp.Category.HelpURL, commandHelp.Category.Name))
	} else {
		text.WriteString(fmt.Sprintf("*%s*", commandHelp.Category.Name))
	}

	if commandHelp.Category.Description != "" {
		text.WriteString(fmt.Sprintf(" (_%s_)", commandHelp.Category.Description))
	}

	text.WriteString(":\n")
}

// ShowSingleCommand prints details of a specific command
func (t *helpCommand) ShowSingleCommand(match matcher.Result, message msg.Message) {
	// compile help only once
	t.once.Do(t.prebuildHelp)

	command := strings.TrimSpace(match.GetString("command"))

	commandHelp, ok := t.commandHelp[command]
	if !ok {
		t.slackClient.SendMessage(message, fmt.Sprintf("Invalid command: `%s`", command))
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

	t.slackClient.SendMessage(message, text)
}

// generate the list of all commands only once and sort them by category/name
func (t *helpCommand) prebuildHelp() {
	var allCommands []bot.Help
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
