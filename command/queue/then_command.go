package queue

import (
	"fmt"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/client"
	log "github.com/sirupsen/logrus"
)

// NewQueueCommand is able to execute a command when another blocking process is done
// e.g. have a running jenkins job and using "then reply done!" to get a information later
func NewQueueCommand(base bot.BaseCommand) bot.Command {
	executeFallbackCommand()

	return &thenCommand{
		base,
	}
}

type thenCommand struct {
	bot.BaseCommand
}

func (c *thenCommand) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher("(?i:queue|then) (?P<command>.*)", c.run)
}

func (c *thenCommand) run(match matcher.Result, message msg.Message) {
	runningCommand, found := runningCommands[message.GetUniqueKey()]
	if !found {
		c.ReplyError(
			message,
			fmt.Errorf("you have to call this command when another long running command is already running"),
		)
		return
	}

	command := match.GetString("command")
	c.AddReaction(waitIcon, message)

	go func() {
		runningCommand.Wait()

		c.AddReaction(doneIcon, message)

		// trigger new command
		client.HandleMessage(message.WithText(command))

		log.Infof("[Queue] Blocking command is over, eval message: %s", command)
	}()
}

func (c *thenCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "then (or queue)",
			Description: "queue a command which is executed when the current task is done",
			Examples: []string{
				"queue reply My job is ready",
				"queue trigger job Deploy master",
				"then trigger job IntegrationTest",
			},
		},
	}
}
