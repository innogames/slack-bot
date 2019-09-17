package jenkins

import (
	"fmt"
	"github.com/bndr/gojenkins"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/client/jenkins"
	"github.com/nlopes/slack"
	"text/template"
)

const (
	actionEnable = "enable"
)

type statusCommand struct {
	jenkins     jenkins.Client
	slackClient client.SlackClient
	jobs        config.JenkinsJobs
}

// newStatusCommand is able to enable/disable (whitelisted) Jenkins jobs
func newStatusCommand(jenkins jenkins.Client, slackClient client.SlackClient, jobs config.JenkinsJobs) bot.Command {
	return &statusCommand{jenkins, slackClient, jobs}
}

func (c *statusCommand) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher("(?P<action>enable|disable) job (?P<job>[\\w\\-_]+)", c.Run)
}

func (c *statusCommand) IsEnabled() bool {
	return c.jenkins != nil
}

func (c *statusCommand) Run(match matcher.Result, event slack.MessageEvent) {
	action := match.GetString("action")
	jobName := match.GetString("job")

	if _, ok := c.jobs[jobName]; !ok {
		message := fmt.Sprintf(
			"Sorry, job *%s* is not whitelisted",
			jobName,
		)
		c.slackClient.Reply(event, message)
		return
	}

	job, err := c.jenkins.GetJob(jobName)
	if err != nil {
		c.slackClient.ReplyError(event, err)
		return
	}

	var message string
	if action == actionEnable {
		_, err = job.Enable()
		message = fmt.Sprintf("Job *%s* is enabled now", jobName)
	} else {
		_, err = job.Disable()
		message = fmt.Sprintf("Job *%s* is disabled now", jobName)
	}

	if err != nil {
		c.slackClient.ReplyError(event, err)
		return
	}

	c.slackClient.Reply(event, message)
}

func (c *statusCommand) GetTemplateFunction() template.FuncMap {
	return template.FuncMap{
		"jenkinsJob": func(jobName string) *gojenkins.Job {
			job, _ := c.jenkins.GetJob(jobName)

			return job
		},
	}
}

func (c *statusCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			"enable job",
			"enabled a jenkins job",
			[]string{
				"enable job MyJobName",
			},
		},
		{
			"disable job",
			"disable a jenkins job",
			[]string{
				"disable job MyJobName",
			},
		},
	}
}
