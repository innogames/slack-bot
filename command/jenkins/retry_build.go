package jenkins

import (
	"context"
	"fmt"
	"net/url"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/command/jenkins/client"
)

type retryCommand struct {
	jenkinsCommand
	jobs config.JenkinsJobs
}

// newRetryCommand initialize a new command to trigger for whitelisted jenkins job
func newRetryCommand(
	base jenkinsCommand,
	jobs config.JenkinsJobs,
) bot.Command {
	return &retryCommand{base, jobs}
}

func (c *retryCommand) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher(`retry (job|build) (?P<job>[\w\-_\\/%.]+)( #?(?P<build>\d+))?`, c.run)
}

func (c *retryCommand) run(match matcher.Result, message msg.Message) {
	jobName := match.GetString("job")

	// URL decode the job name to handle multibranch pipeline names with encoded characters
	decodedJobName, err := url.QueryUnescape(jobName)
	if err != nil {
		// If decoding fails, use the original job name
		decodedJobName = jobName
	}

	if _, ok := c.jobs[decodedJobName]; !ok {
		c.ReplyError(message, fmt.Errorf("job *%s* is not whitelisted", decodedJobName))
		return
	}

	ctx := context.TODO()
	job, err := c.jenkins.GetJob(ctx, decodedJobName)
	if err != nil {
		c.SendMessage(message, fmt.Sprintf("Job *%s* does not exist", decodedJobName))
		return
	}

	buildNumber := match.GetInt("build")
	build, err := getBuild(ctx, job, buildNumber)
	if err != nil {
		c.ReplyError(message, fmt.Errorf("given build *%s #%d* does not exist: %w", decodedJobName, buildNumber, err))
		return
	}

	parameters := make(client.Parameters)
	for _, param := range build.GetParameters() {
		parameters[param.Name] = param.Value
	}

	err = client.TriggerJenkinsJob(c.jobs[decodedJobName], decodedJobName, parameters, c.SlackClient, c.jenkins, message)
	if err != nil {
		c.ReplyError(message, err)
	}
}

func (c *retryCommand) GetHelp() []bot.Help {
	examples := []string{
		"retry job BuildSomeJob",
		"retry job BuildSomeJob #101",
	}

	var help []bot.Help
	help = append(help, bot.Help{
		Command:     "retry job <job> [<build>]",
		Description: "restart the most recent jenkins build of the given job",
		Examples:    examples,
		Category:    category,
	})

	return help
}
