package jenkins

import (
	"context"
	"time"

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
	return matcher.NewTextMatcher("wait until jenkins is idle", c.run)
}

func (c *idleWatcherCommand) run(match matcher.Result, message msg.Message) {
	if !c.hasRunningBuild(message) {
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
			if c.hasRunningBuild(message) {
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
func (c *idleWatcherCommand) hasRunningBuild(ref msg.Ref) bool {
	ctx := context.TODO()
	nodes, err := c.jenkins.GetAllNodes(ctx)
	if err != nil {
		c.ReplyError(ref, err)
		return false
	}
	for _, node := range nodes {
		for _, executor := range node.Raw.Executors {
			if executor.CurrentExecutable.Number != 0 {
				// there is something running!
				return true
			}
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
	}
}
