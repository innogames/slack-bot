package jenkins

import (
	"context"
	"fmt"
	"text/template"

	"github.com/bndr/gojenkins"
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
)

const (
	actionEnable = "enable"
)

type statusCommand struct {
	jenkinsCommand
	jobs config.JenkinsJobs
}

// newStatusCommand is able to enable/disable (whitelisted) Jenkins jobs
func newStatusCommand(base jenkinsCommand, jobs config.JenkinsJobs) bot.Command {
	return &statusCommand{base, jobs}
}

func (c *statusCommand) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher(`(?P<action>enable|disable) job (?P<job>[\w\-_\\/]+)`, c.run)
}

func (c *statusCommand) run(match matcher.Result, message msg.Message) {
	action := match.GetString("action")
	jobName := match.GetString("job")

	ctx := context.TODO()

	if _, ok := c.jobs[jobName]; !ok {
		text := fmt.Sprintf(
			"Sorry, job *%s* is not whitelisted",
			jobName,
		)
		c.SendMessage(message, text)
		return
	}

	job, err := c.jenkins.GetJob(ctx, jobName)
	if err != nil {
		c.ReplyError(message, err)
		return
	}

	var text string

	if action == actionEnable {
		_, err = job.Enable(ctx)
		text = fmt.Sprintf("Job *%s* is enabled now", jobName)
	} else {
		_, err = job.Disable(ctx)
		text = fmt.Sprintf("Job *%s* is disabled now", jobName)
	}

	if err != nil {
		c.ReplyError(message, err)
		return
	}

	c.SendMessage(message, text)
}

func (c *statusCommand) GetTemplateFunction() template.FuncMap {
	return template.FuncMap{
		"jenkinsJob": func(jobName string) *gojenkins.Job {
			job, _ := c.jenkins.GetJob(context.TODO(), jobName)

			return job
		},
	}
}

func (c *statusCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "enable job <job>",
			Description: "enabled a jenkins job",
			Examples: []string{
				"enable job MyJobName",
			},
			Category: category,
		},
		{
			Command:     "disable job <job>",
			Description: "disable a jenkins job",
			Examples: []string{
				"disable job MyJobName",
			},
			Category: category,
		},
	}
}
