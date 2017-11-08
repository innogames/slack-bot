package jenkins

import (
	"fmt"
	"github.com/bndr/gojenkins"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/client/jenkins"
	"github.com/innogames/slack-bot/command/queue"
	"github.com/nlopes/slack"
)

type buildWatcherCommand struct {
	jenkins     jenkins.Client
	slackClient client.SlackClient
}

// NewBuildWatcherCommand watches the status of an already running jenkins build
func NewBuildWatcherCommand(jenkins jenkins.Client, slackClient client.SlackClient) bot.Command {
	return &buildWatcherCommand{jenkins, slackClient}
}

func (c *buildWatcherCommand) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher(`(notify|inform)( me about)? (job|build) ?(?P<job>[\w\-_]*)( #?(?P<build>\d+))?`, c.Run)
}

func (c *buildWatcherCommand) IsEnabled() bool {
	return c.jenkins != nil
}

func (c *buildWatcherCommand) Run(match matcher.Result, event slack.MessageEvent) {
	jobName := match.GetString("job")
	buildNumber := match.GetInt("build")

	job, err := c.jenkins.GetJob(jobName)
	if err != nil {
		c.slackClient.Reply(event, fmt.Sprintf("Job %s does not exist", jobName))
		return
	}

	build, err := getBuild(job, buildNumber)
	if err != nil {
		c.slackClient.ReplyError(event, err)
		return
	}

	if !build.IsRunning() {
		c.slackClient.Reply(event, fmt.Sprintf("No job for *%s* is running right now", jobName))
		return
	}

	msg := fmt.Sprintf(
		"Okay, I'll inform you when the job %s #%s is done",
		jobName,
		build.Info().ID,
	)
	attachment := jenkins.GetAttachment(build, msg)
	msgTimestamp := c.slackClient.SendMessage(event, "", slack.MsgOptionAttachments(attachment))
	newMsgRef := slack.NewRefToMessage(event.Channel, msgTimestamp)

	done := queue.AddRunningCommand(
		event,
		fmt.Sprintf("inform job %s #%d", jobName, build.GetBuildNumber()),
	)
	go func() {
		<-jenkins.WatchBuild(build)
		done <- true

		attachment = jenkins.GetAttachment(build, msg)
		c.slackClient.SendMessage(
			event,
			"",
			slack.MsgOptionUpdate(msgTimestamp),
			slack.MsgOptionAttachments(attachment),
		)

		c.slackClient.RemoveReaction(jenkins.IconRunning, newMsgRef)
		if build.IsGood() {
			c.slackClient.AddReaction(jenkins.IconSuccess, newMsgRef)
		} else {
			c.slackClient.AddReaction(jenkins.IconFailed, newMsgRef)
		}

		c.slackClient.Reply(event, fmt.Sprintf(
			"<@%s> *%s*: %s #%s: %s",
			event.User,
			build.GetResult(),
			jobName,
			build.Info().ID,
			build.GetUrl()),
		)
	}()
}

func getBuild(job jenkins.Job, buildNumber int) (*gojenkins.Build, error) {
	if buildNumber == 0 {
		job.Poll()
		return job.GetLastBuild()
	} else {
		return job.GetBuild(int64(buildNumber))
	}
}

func (c *buildWatcherCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			"notify build",
			"inform you when a running jenkins build finishes",
			[]string{
				"inform me about build AtcBrowser #1233",
				"inform me about build AtcMobile",
				"notify build AtcMobile",
			},
		},
		{
			"inform job",
			"inform you when a running jenkins build finishes",
			[]string{
				"inform me about build AtcBrowser #1233",
				"inform me about build AtcMobile",
				"notify build AtcMobile",
			},
		},
	}
}
