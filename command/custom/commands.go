package custom

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/storage"
	"github.com/innogames/slack-bot/client"
	"github.com/nlopes/slack"
)

const storeKey = "user_commands"

type baseCommand struct {
	slackClient client.SlackClient
}

type list map[string]string

func loadList(event slack.MessageEvent) list {
	list := make(list, 0)

	storage.Read(storeKey, event.User, &list)

	return list
}

func storeList(event slack.MessageEvent, list list) {
	storage.Write(storeKey, event.User, list)
}

func (c *baseCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			"custom commands",
			"Define command aliases which are just available for you. You can use a `;` to separate single commands",
			[]string{
				"`list commands`",
				"`add command 'myCommand' as 'trigger job RestoreWorld 7'` -> then just call `myCommand` later",
				"`add command 'build master' 'trigger job Deploy master ; then trigger job DeployClient master'`",
				"`delete command 'build master'`",
			},
		},
	}
}

// GetCommands returns a set of all custom commands
// todo: use one command with a GroupMatcher
func GetCommands(slackClient client.SlackClient) bot.Commands {
	var commands bot.Commands

	commands = bot.Commands{}
	commands.AddCommand(
		&addCommand{baseCommand{slackClient}},
		&deleteCommand{baseCommand{slackClient}},
		&listCommand{baseCommand{slackClient}},
		&handleCommand{baseCommand{slackClient}},
	)

	return commands
}
