package jenkins

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/client/jenkins"
	"github.com/slack-go/slack"
	"github.com/sirupsen/logrus"
)

type retryCommand struct {
	jenkins     jenkins.Client
	slackClient client.SlackClient
	jobs        config.JenkinsJobs
	logger      *logrus.Logger
}

// newRetryCommand initialize a new command to trigger for whitelisted jenkins job
func newRetryCommand(
	jenkins jenkins.Client,
	slackClient client.SlackClient,
	jobs config.JenkinsJobs,
	logger *logrus.Logger,
) bot.Command {
	return &retryCommand{jenkins, slackClient, jobs, logger}
}

func (c *retryCommand) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher("retry (job|build) (?P<job>[\\w\\-_]+)( #?(?P<build>\\d+))?", c.Run)
}

func (c *retryCommand) IsEnabled() bool {
	return c.jenkins != nil
}

func (c *retryCommand) Run(match matcher.Result, event slack.MessageEvent) {
	jobName := match.GetString("job")
	if _, ok := c.jobs[jobName]; !ok {
		c.slackClient.ReplyError(event, fmt.Errorf("job *%s* is not whitelisted", jobName))
		return
	}

	job, err := c.jenkins.GetJob(jobName)
	if err != nil {
		c.slackClient.Reply(event, fmt.Sprintf("Job *%s* does not exist", jobName))
		return
	}

	buildNumber := match.GetInt("build")
	build, err := getBuild(job, buildNumber)
	if err != nil {
		c.slackClient.ReplyError(event, fmt.Errorf("Given build *%s #%d* does not exist: %s", jobName, buildNumber, err.Error()))
		return
	}

	parameters := make(jenkins.Parameters)
	for _, param := range build.GetParameters() {
		parameters[param.Name] = param.Value
	}

	err = jenkins.TriggerJenkinsJob(c.jobs[jobName], jobName, parameters, c.slackClient, c.jenkins, event, c.logger)
	if err != nil {
		c.slackClient.ReplyError(event, err)
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
		Description: "restart the most recent jenkins build of the givenn job",
		Examples:    examples,
	})

	return help
}
