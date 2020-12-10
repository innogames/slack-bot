package queue

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/client"
	log "github.com/sirupsen/logrus"
	"time"
)

// NewQueueCommand is able to execute a command when another blocking process is done
// e.g. have a running jenkins job and using "then reply done!" to get a information later
func NewQueueCommand(base bot.BaseCommand) bot.Command {
	executeFallbackCommand()

	return &command{
		base,
	}
}

type command struct {
	bot.BaseCommand
}

func (c *command) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher("(?i:queue|then) (?P<command>.*)", c.Run)
}

func (c *command) Run(match matcher.Result, message msg.Message) {
	if !IsBlocked(message) {
		c.ReplyError(
			message,
			fmt.Errorf("you have to call this command when another long running command is already running"),
		)
		return
	}

	command := match.GetString("command")
	c.AddReaction(waitIcon, message)

	key := message.GetUniqueKey()

	go func() {
		// todo avoid polling here by another chan etc
		ticker := time.NewTicker(time.Millisecond * 250)
		defer ticker.Stop()

		for range ticker.C {
			mu.RLock()
			if _, ok := runningCommands[key]; ok {
				// still running...
				mu.RUnlock()

				continue
			}
			mu.RUnlock()
			c.AddReaction(doneIcon, message)

			// trigger new command
			client.InternalMessages <- message.WithText(command)

			log.Infof("[Queue] Blocking command is over, eval newMessage: %s", command)
			return
		}
	}()
}

func (c *command) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "queue",
			Description: "queue a command which is executed when the current task is done",
			Examples: []string{
				"queue reply My job is ready",
				"queue trigger job Deploy master",
				"then trigger job IntegrationTest",
			},
		},
	}
}
