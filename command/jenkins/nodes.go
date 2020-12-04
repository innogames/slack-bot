package jenkins

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/bot/msg"
)

const (
	iconStatusOnline  = ":check_mark:"
	iconStatusOffline = ":red_circle:"
)

type nodesCommand struct {
	jenkinsCommand
}

// newNodesCommand lists all Jenkins nodes/slaves and the current number of running executors
func newNodesCommand(base jenkinsCommand) bot.Command {
	return &nodesCommand{base}
}

func (c *nodesCommand) GetMatcher() matcher.Matcher {
	return matcher.NewTextMatcher("jenkins nodes", c.Run)
}

func (c *nodesCommand) IsEnabled() bool {
	return c.jenkins != nil
}

func (c *nodesCommand) Run(match matcher.Result, message msg.Message) {
	nodes, err := c.jenkins.GetAllNodes()
	if err != nil {
		c.ReplyError(message, err)
		return
	}

	text := fmt.Sprintf("*%d Nodes*\n", len(nodes))
	var statusIcon string
	for _, node := range nodes {
		offline := node.Raw.Offline

		if offline {
			statusIcon = iconStatusOffline
		} else {
			statusIcon = iconStatusOnline
		}

		text += fmt.Sprintf(
			"- *%s* - status: %s - executors: %d\n",
			node.GetName(),
			statusIcon,
			len(node.Raw.Executors),
		)
	}

	c.SendMessage(message, text)
}

func (c *nodesCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "jenkins nodes",
			Description: "Prints a list of all jenkins nodes",
			Examples: []string{
				"jenkins nodes",
			},
			Category: category,
		},
	}
}
