package jenkins

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
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
	host string // configured Jenkins host for URL validation
}

const (
	iconRunning = "arrows_counterclockwise"
	iconSuccess = "white_check_mark"
	iconFailed  = "x"
)

// newBuildWatcherCommand watches the status of an already running jenkins build
func newBuildWatcherCommand(base jenkinsCommand, host string) bot.Command {
	return &buildWatcherCommand{base, host}
}

func (c *buildWatcherCommand) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher(`(notify|inform)( me about)? (job|build) ?(?P<job>[\w\-_\\/%.]+|https://[^\s]+)( #?(?P<build>\d+))?`, c.watch)
}

func (c *buildWatcherCommand) watch(match matcher.Result, message msg.Message) {
	jobName := match.GetString("job")
	buildNumber := match.GetInt("build")

	// Check if input is a valid build URL
	if strings.HasPrefix(jobName, "https://") {
		// Validate host matches configured host
		parsedURL, err := url.Parse(jobName)
		if err != nil || !strings.HasPrefix(c.host, "https://"+parsedURL.Host) {
			c.SendMessage(message, "URL does not match configured Jenkins host")
			return
		}
		// Parse URL to extract job name and build number
		var parseErr error
		jobName, buildNumber, parseErr = parseJenkinsURL(jobName)
		if parseErr != nil {
			c.SendMessage(message, fmt.Sprintf("Invalid Jenkins URL: %s", parseErr))
			return
		}
	}

	// URL decode the job name to handle multibranch pipeline names with encoded characters
	decodedJobName, err := url.QueryUnescape(jobName)
	if err != nil {
		// If decoding fails, use the original job name
		decodedJobName = jobName
	}

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

// parseJenkinsURL extracts job name and build number from a Jenkins URL
// URL format: https://host/job/JobName/BuildNumber/ or
//
//	https://host/job/Folder/job/JobName/BuildNumber/
func parseJenkinsURL(urlStr string) (jobName string, buildNumber int, err error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", 0, err
	}

	// Use RawPath to preserve URL-encoded characters (e.g., %2F in branch names)
	// Fall back to Path if RawPath is empty
	path := parsedURL.RawPath
	if path == "" {
		path = parsedURL.Path
	}

	// Remove trailing slash and split path
	path = strings.TrimSuffix(path, "/")
	segments := strings.Split(path, "/")

	// Filter out empty segments
	var filteredSegments []string
	for _, seg := range segments {
		if seg != "" {
			filteredSegments = append(filteredSegments, seg)
		}
	}

	if len(filteredSegments) == 0 {
		return "", 0, errors.New("invalid Jenkins URL: no path segments")
	}

	// Parse the path to extract job names and build number
	// Pattern: /job/X/job/Y/.../BuildNumber or /job/X/job/Y/...
	var jobParts []string
	for i := 0; i < len(filteredSegments); i++ {
		if filteredSegments[i] == "job" && i+1 < len(filteredSegments) {
			// Next segment is a job/folder name
			jobParts = append(jobParts, filteredSegments[i+1])
			i++ // Skip the job name segment
		} else {
			// This might be a build number (last non-"job" segment)
			if num, parseErr := strconv.Atoi(filteredSegments[i]); parseErr == nil {
				buildNumber = num
			}
		}
	}

	if len(jobParts) == 0 {
		return "", 0, errors.New("invalid Jenkins URL: no job found in path")
	}

	jobName = strings.Join(jobParts, "/")
	return jobName, buildNumber, nil
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
