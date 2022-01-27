package jenkins

import (
	"context"
	"fmt"
	"sort"

	"github.com/bndr/gojenkins"
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
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
	return matcher.NewTextMatcher("list jenkins nodes", c.run)
}

func (c *nodesCommand) run(match matcher.Result, message msg.Message) {
	ctx := context.TODO()
	nodes, err := c.jenkins.GetAllNodes(ctx)
	if err != nil {
		c.ReplyError(message, err)
		return
	}

	// sort nodes by name
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].GetName() < nodes[j].GetName()
	})

	text := fmt.Sprintf("*<%s/computer/|%d Nodes>*\n", c.cfg.Host, len(nodes))

	totalJobsRunning := 0

	for _, node := range nodes {
		runningJobs := countBusyExecutors(node)
		totalJobsRunning += runningJobs

		text += fmt.Sprintf(
			"• *<%s/computer/%s/|%s>* - %s - busy executors: %d/%d\n",
			c.cfg.Host,
			node.GetName(),
			node.GetName(),
			getNodeStatus(node),
			runningJobs,
			len(node.Raw.Executors)+len(node.Raw.OneOffExecutors),
		)
	}

	text += fmt.Sprintf("\nIn total there are %d build(s) running right now", totalJobsRunning)

	c.SendMessage(message, text)
}

func getNodeStatus(node *gojenkins.Node) string {
	if node.Raw.Offline {
		return "offline 🔴"
	}
	if node.Raw.TemporarilyOffline {
		return "temporary offline ⏸"
	}

	return "online ✔"
}

func countBusyExecutors(node *gojenkins.Node) int {
	busyExecutors := len(node.Raw.OneOffExecutors)

	for _, executor := range node.Raw.Executors {
		if executor.CurrentExecutable.Number != 0 {
			busyExecutors++
		}
	}

	return busyExecutors
}

func (c *nodesCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "list jenkins nodes",
			Description: "Prints a list of all jenkins nodes",
			Category:    category,
		},
	}
}
