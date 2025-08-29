package jenkins

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/bndr/gojenkins"
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/innogames/slack-bot/v2/command/jenkins/client"
	"github.com/innogames/slack-bot/v2/command/queue"
	"github.com/slack-go/slack"
)

type buildWatcherCommand struct {
	jenkinsCommand
}

const (
	iconRunning = "arrows_counterclockwise"
	iconSuccess = "white_check_mark"
	iconFailed  = "x"
)

// newBuildWatcherCommand watches the status of an already running jenkins build
func newBuildWatcherCommand(base jenkinsCommand) bot.Command {
	return &buildWatcherCommand{base}
}

func (c *buildWatcherCommand) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher(`(notify|inform)( me about)? (job|build) ?(?P<job>[\w\-_\\/%.]+)( #?(?P<build>\d+))?`, c.watch)
}

func (c *buildWatcherCommand) watch(match matcher.Result, message msg.Message) {
	jobName := match.GetString("job")

	// URL decode the job name to handle multibranch pipeline names with encoded characters
	decodedJobName, err := url.QueryUnescape(jobName)
	if err != nil {
		// If decoding fails, use the original job name
		decodedJobName = jobName
	}

	buildNumber := match.GetInt("build")

	ctx := context.Background()
	job, err := c.jenkins.GetJob(ctx, decodedJobName)
	if err != nil {
		c.SendMessage(message, fmt.Sprintf("Job *%s* does not exist", decodedJobName))
		return
	}

	build, err := getBuild(ctx, job, buildNumber)
	if err != nil {
		c.SendMessage(message, fmt.Sprintf("Build *%s#%d* does not exist", decodedJobName, buildNumber))
		return
	}

	if !build.Raw.Building {
		c.SendMessage(message, fmt.Sprintf("No job for *%s* is running right now", decodedJobName))
		return
	}

	text := fmt.Sprintf(
		"Okay, I'll inform you when the job %s #%s is done",
		decodedJobName,
		build.Info().ID,
	)
	attachment := client.GetAttachment(build, text)
	msgTimestamp := c.SendMessage(message, "", attachment)

	runningCommand := queue.AddRunningCommand(
		message,
		fmt.Sprintf("inform job %s #%d", decodedJobName, build.GetBuildNumber()),
	)

	c.AddReaction(iconRunning, message)
	go func() {
		<-client.WatchBuild(build)
		runningCommand.Done()

		c.SendMessage(
			message,
			"",
			slack.MsgOptionUpdate(msgTimestamp),
			client.GetAttachment(build, text),
		)

		c.RemoveReaction(iconRunning, message)
		if build.IsGood(ctx) {
			c.AddReaction(iconSuccess, message)
		} else {
			c.AddReaction(iconFailed, message)
		}

		duration := time.Duration(build.GetDuration()) * time.Millisecond
		c.SendMessage(message, fmt.Sprintf(
			"<@%s> *%s*: %s #%s: %s in %s",
			message.User,
			build.GetResult(),
			decodedJobName,
			build.Info().ID,
			build.GetUrl(),
			util.FormatDuration(duration),
		))
	}()
}

func getBuild(ctx context.Context, job client.Job, buildNumber int) (*gojenkins.Build, error) {
	if buildNumber == 0 {
		_, err := job.Poll(ctx)
		if err != nil {
			return nil, err
		}

		return job.GetLastBuild(ctx)
	}
	return job.GetBuild(ctx, int64(buildNumber))
}

func (c *buildWatcherCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "notify build <job> [<build>]",
			Description: "inform you when a running jenkins build finishes",
			Examples: []string{
				"inform me about build AtcBrowser #1233",
				"inform me about build AtcMobile",
				"notify build AtcMobile",
			},
			Category: category,
		},
		{
			Command:     "inform <job> [<build>]",
			Description: "inform you when a running jenkins build finishes",
			Examples: []string{
				"inform me about build AtcBrowser #1233",
				"inform me about build AtcMobile",
				"notify build AtcMobile",
			},
			Category: category,
		},
	}
}
