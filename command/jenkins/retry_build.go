package jenkins

import (
	"context"
	"fmt"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/client/jenkins"
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
	return matcher.NewRegexpMatcher(`retry (job|build) (?P<job>[\w\-_\\/]+)( #?(?P<build>\d+))?`, c.run)
}

func (c *retryCommand) run(match matcher.Result, message msg.Message) {
	jobName := match.GetString("job")
	if _, ok := c.jobs[jobName]; !ok {
		c.ReplyError(message, fmt.Errorf("job *%s* is not whitelisted", jobName))
		return
	}

	ctx := context.TODO()
	job, err := c.jenkins.GetJob(ctx, jobName)
	if err != nil {
		c.SendMessage(message, fmt.Sprintf("Job *%s* does not exist", jobName))
		return
	}

	buildNumber := match.GetInt("build")
	build, err := getBuild(ctx, job, buildNumber)
	if err != nil {
		c.ReplyError(message, fmt.Errorf("given build *%s #%d* does not exist: %w", jobName, buildNumber, err))
		return
	}

	parameters := make(jenkins.Parameters)
	for _, param := range build.GetParameters() {
		parameters[param.Name] = param.Value
	}

	err = jenkins.TriggerJenkinsJob(c.jobs[jobName], jobName, parameters, c.SlackClient, c.jenkins, message)
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
