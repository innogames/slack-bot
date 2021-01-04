package jenkins

import (
	"fmt"

	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/client/jenkins"
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
	return matcher.NewRegexpMatcher("retry (job|build) (?P<job>[\\w\\-_]+)( #?(?P<build>\\d+))?", c.Run)
}

func (c *retryCommand) IsEnabled() bool {
	return c.jenkins != nil
}

func (c *retryCommand) Run(match matcher.Result, message msg.Message) {
	jobName := match.GetString("job")
	if _, ok := c.jobs[jobName]; !ok {
		c.ReplyError(message, fmt.Errorf("job *%s* is not whitelisted", jobName))
		return
	}

	job, err := c.jenkins.GetJob(jobName)
	if err != nil {
		c.SendMessage(message, fmt.Sprintf("Job *%s* does not exist", jobName))
		return
	}

	buildNumber := match.GetInt("build")
	build, err := getBuild(job, buildNumber)
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
		Command:     "retry job",
		Description: "restart the most recent jenkins build of the given job",
		Examples:    examples,
		Category:    category,
	})

	return help
}
