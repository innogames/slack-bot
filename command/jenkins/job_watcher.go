package jenkins

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/client/jenkins"
	"github.com/slack-go/slack"
	"github.com/sirupsen/logrus"
)

const (
	actionWatch   = "watch"
	actionUnwatch = "unwatch"
)

// newJobWatcherCommand initialize a new command to watch for any jenkins job
func newJobWatcherCommand(jenkinsClient jenkins.Client, slackClient client.SlackClient, logger *logrus.Logger) bot.Command {
	return &watcherCommand{
		jenkinsClient,
		slackClient,
		logger,
		make(map[string]chan bool, 0),
	}
}

type watcherCommand struct {
	jenkins     jenkins.Client
	slackClient client.SlackClient
	logger      *logrus.Logger
	stopper     map[string]chan bool
}

func (c *watcherCommand) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher(`(?P<action>watch|unwatch) (?P<job>[\w\-_]+)`, c.Run)
}

func (c *watcherCommand) IsEnabled() bool {
	return c.jenkins != nil
}

func (c *watcherCommand) Run(match matcher.Result, event slack.MessageEvent) {
	action := match.GetString("action")
	jobName := match.GetString("job")
	if action == actionWatch {
		stop := make(chan bool, 1)
		c.stopper[jobName+event.User] = stop
		builds, err := jenkins.WatchJob(c.jenkins, jobName, stop)
		if err != nil {
			c.slackClient.ReplyError(event, err)
			return
		}
		c.slackClient.Reply(event, fmt.Sprintf("Okay, I'll watch %s\nUnwatch via `unwatch %s`", jobName, jobName))

		go func() {
			for build := range builds {
				c.slackClient.Reply(event, fmt.Sprintf(
					"*%s*: %s #%d: %s",
					build.GetResult(),
					jobName,
					build.GetBuildNumber(),
					build.GetUrl()),
				)
			}
		}()
	}

	if action == actionUnwatch {
		if stop, ok := c.stopper[jobName+event.User]; ok {
			stop <- true
		}
		c.slackClient.Reply(event, fmt.Sprintf("Okay, you just unwatched %s", jobName))
	}
}

func (c *watcherCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			"watch job",
			"watch jenkins jobs and informs about successful/error jobs",
			[]string{
				"watch MyJobName",
				"unwatch MyJobName",
			},
		},
	}
}
