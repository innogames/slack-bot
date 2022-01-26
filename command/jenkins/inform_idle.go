package jenkins

import (
	"context"
	"time"

	"github.com/bndr/gojenkins"
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/command/queue"
)

type idleWatcherCommand struct {
	jenkinsCommand
	checkInterval time.Duration
}

const (
	idleCheckInterval = time.Minute * 1

	waitingReaction = "coffee"
	doneReaction    = "white_check_mark"
)

func newIdleWatcherCommand(base jenkinsCommand) bot.Command {
	return &idleWatcherCommand{base, idleCheckInterval}
}

func (c *idleWatcherCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewTextMatcher("wait until jenkins is idle", c.checkAllNodes),
		matcher.NewRegexpMatcher(`wait until jenkins node (?P<node>\w+) is idle`, c.checkSingleNode),
	)
}

// command like "wait until node swarm-1234 is idle"
func (c *idleWatcherCommand) checkSingleNode(match matcher.Result, message msg.Message) {
	nodeName := match.GetString("node")

	filter := func(node *gojenkins.Node) bool {
		return node.GetName() == nodeName
	}
	c.check(message, filter)
}

// command like "wait until jenkins is idle"
func (c *idleWatcherCommand) checkAllNodes(match matcher.Result, message msg.Message) {
	filter := func(node *gojenkins.Node) bool { return true }
	c.check(message, filter)
}

func (c *idleWatcherCommand) check(message msg.Message, nodeFilter func(node *gojenkins.Node) bool) {
	if !c.hasRunningBuild(message, nodeFilter) {
		c.AddReaction(doneReaction, message)
		c.SendMessage(
			message,
			"There are no jobs running right now!",
		)
		return
	}

	runningCommand := queue.AddRunningCommand(
		message,
		message.Text,
	)
	c.AddReaction(waitingReaction, message)

	go func() {
		timer := time.NewTicker(c.checkInterval)
		defer timer.Stop()

		for range timer.C {
			if c.hasRunningBuild(message, nodeFilter) {
				// still builds running...
				continue
			}

			c.SendMessage(
				message,
				"No job is running anymore",
			)
			c.RemoveReaction(waitingReaction, message)
			c.AddReaction(doneReaction, message)

			// mark queued command as done to perform next "then" command
			runningCommand.Done()

			return
		}
	}()
}

// query all executors from jenkins with one request and check of any executor is busy
func (c *idleWatcherCommand) hasRunningBuild(ref msg.Ref, nodeFilter func(build *gojenkins.Node) bool) bool {
	ctx := context.Background()
	nodes, err := c.jenkins.GetAllNodes(ctx)
	if err != nil {
		c.ReplyError(ref, err)
		return false
	}

	for _, node := range nodes {
		if !nodeFilter(node) {
			// current command is not interested in this node...
			continue
		}

		if countBusyExecutors(node) > 0 {
			// there is something running!
			return true
		}
	}

	return false
}

func (c *idleWatcherCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "wait until jenkins is idle",
			Description: "Informs you if no Jenkins job is running anymore. Useful when we're planning updates/maintenance which requires a idle server.",
			Examples: []string{
				"wait until jenkins is idle",
			},
			Category: category,
		},
		{
			Command:     "wait until node <nodeId> is idle",
			Description: "Informs you if no Jenkins job is running on the given node anymore. Useful when we're planning updates/maintenance which requires a idle server.",
			Examples: []string{
				"wait until jenkins is idle",
				"wait until node swarm-1234 is idle",
			},
			Category: category,
		},
	}
}
