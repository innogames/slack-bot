package queue

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/nlopes/slack"
	"github.com/sirupsen/logrus"
	"time"
)

// NewQueueCommand is able to execute a command when another blocking process is done
// e.g. have a running jenkins job and using "then reply done!" to get a information later
func NewQueueCommand(slackClient client.SlackClient, log *logrus.Logger) bot.Command {
	logger = log

	executeFallbackCommand(logger)

	return &command{
		slackClient,
		logger,
	}
}

type command struct {
	slackClient client.SlackClient
	logger      *logrus.Logger
}

func (c *command) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher("(?i:queue|then) (?P<command>.*)", c.Run)
}

func (c *command) Run(match matcher.Result, event slack.MessageEvent) {
	if !IsBlocked(event) {
		c.slackClient.ReplyError(
			event,
			fmt.Errorf("You have to call this command when another long running command is already running"),
		)
		return
	}

	command := match.GetString("command")
	msgRef := slack.NewRefToMessage(event.Channel, event.Timestamp)
	c.slackClient.AddReaction(waitIcon, msgRef)

	go func() {
		// todo avoid polling here by another chan etc + make thread safe
		ticker := time.NewTicker(time.Millisecond * 250)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			if _, ok := runningCommands[getKey(event)]; ok {
				// still running...
				mu.Unlock()

				continue
			}
			mu.Unlock()
			c.slackClient.AddReaction(doneIcon, msgRef)

			// trigger new command
			newMessage := event
			newMessage.Text = command
			client.InternalMessages <- newMessage

			c.logger.Infof("[Queue] Blocking command is over, eval newMessage: %s", newMessage.Text)
			return
		}
	}()
}

func (c *command) GetHelp() []bot.Help {
	return []bot.Help{
		{
			"queue",
			"queue a queuedCommand which is executed when the current task is done",
			[]string{
				"queue reply My job is ready",
				"queue trigger job Deploy master",
				"then trigger job IntegrationTest",
			},
		},
	}
}
