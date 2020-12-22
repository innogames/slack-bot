package jenkins

import (
	"fmt"
	"github.com/bndr/gojenkins"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/bot/msg"
	"sort"
)

const (
	iconStatusOnline  = ":white_check_mark:"
	iconStatusOffline = ":red_circle:"
)

type nodesCommand struct {
	jenkinsCommand
	cfg config.Jenkins
}

// newNodesCommand lists all Jenkins nodes/slaves and the current number of running executors
func newNodesCommand(base jenkinsCommand, cfg config.Jenkins) bot.Command {
	return &nodesCommand{base, cfg}
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

	// sort nodes by name
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].GetName() < nodes[j].GetName()
	})

	text := fmt.Sprintf("*<%s/computer/|%d Nodes>*\n", c.cfg.Host, len(nodes))
	var statusIcon string
	for _, node := range nodes {
		offline := node.Raw.Offline

		if offline {
			statusIcon = iconStatusOffline
		} else {
			statusIcon = iconStatusOnline
		}

		text += fmt.Sprintf(
			"• *<%s/computer/%s/|%s>* - status: %s - busy executors: %d/%d\n",
			c.cfg.Host,
			node.GetName(),
			node.GetName(),
			statusIcon,
			c.countBusyExecutors(node),
			len(node.Raw.Executors),
		)
	}

	c.SendMessage(message, text)
}

func (c *nodesCommand) countBusyExecutors(node *gojenkins.Node) int {
	busyNodes := 0
	for _, executor := range node.Raw.Executors {
		if executor.CurrentExecutable.Number != 0 {
			busyNodes++
		}
	}

	return busyNodes
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
