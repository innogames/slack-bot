package jenkins

import (
	"context"
	"fmt"
	"net/url"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/command/jenkins/client"
)

const (
	actionWatch   = "watch"
	actionUnwatch = "unwatch"
)

// newJobWatcherCommand initialize a new command to watch for any jenkins job
func newJobWatcherCommand(base jenkinsCommand) bot.Command {
	return &watcherCommand{
		base,
		make(map[string]chan bool),
	}
}

type watcherCommand struct {
	jenkinsCommand
	stopper map[string]chan bool
}

func (c *watcherCommand) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher(`(?P<action>watch|unwatch) (?P<job>[\w\-_\\/%.]+)`, c.run)
}

func (c *watcherCommand) run(match matcher.Result, message msg.Message) {
	action := match.GetString("action")
	jobName := match.GetString("job")

	// URL decode the job name to handle multibranch pipeline names with encoded characters
	decodedJobName, err := url.QueryUnescape(jobName)
	if err != nil {
		// If decoding fails, use the original job name
		decodedJobName = jobName
	}

	if action == actionWatch {
		stop := make(chan bool, 1)
		ctx := context.TODO()
		// todo use context.WithCancel instead of stopper chan
		c.stopper[decodedJobName+message.GetUser()] = stop
		builds, err := client.WatchJob(ctx, c.jenkins, decodedJobName, stop)
		if err != nil {
			c.ReplyError(message, err)
			return
		}

		c.SendMessage(message, fmt.Sprintf("Okay, I'll watch %s\nUnwatch via `unwatch %s`", decodedJobName, decodedJobName))

		go func() {
			for build := range builds {
				c.SendMessage(message, fmt.Sprintf(
					"*%s*: %s #%d: %s",
					build.GetResult(),
					decodedJobName,
					build.GetBuildNumber(),
					build.GetUrl()),
				)
			}
		}()
	}

	if action == actionUnwatch {
		if stop, ok := c.stopper[decodedJobName+message.User]; ok {
			stop <- true
		}

		c.SendMessage(message, "Okay, you just unwatched "+decodedJobName)
	}
}

func (c *watcherCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "watch job <job>",
			Description: "watch jenkins jobs and informs about successful/error jobs",
			Examples: []string{
				"watch MyJobName",
				"unwatch MyJobName",
			},
			Category: category,
		},
		{
			Command:     "unwatch job <job>",
			Description: "unwatch jenkins jobs and informs about successful/error jobs",
			Examples: []string{
				"watch MyJobName",
				"unwatch MyJobName",
			},
			Category: category,
		},
	}
}
