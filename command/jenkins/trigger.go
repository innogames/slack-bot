package jenkins

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/innogames/slack-bot/v2/client/jenkins"
)

// command to trigger/start jenkins jobs
type triggerCommand struct {
	jenkinsCommand
	jobs map[string]triggerCommandData
	cfg  config.JenkinsJobs
}

type triggerCommandData struct {
	jobName string
	config  config.JobConfig
	trigger *regexp.Regexp
}

// newTriggerCommand initialize a new command to trigger for whitelisted jenkins job
func newTriggerCommand(
	base jenkinsCommand,
	jobs config.JenkinsJobs,
) bot.Command {
	trigger := make(map[string]triggerCommandData, len(jobs))

	for jobName, cfg := range jobs {
		trigger[jobName] = triggerCommandData{
			jobName,
			cfg,
			util.CompileRegexp(cfg.Trigger),
		}
	}

	return &triggerCommand{base, trigger, jobs}
}

func (c *triggerCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewRegexpMatcher(`((trigger|start) (jenkins|build|job)) (?P<job>[\w\-_\\/]*)(?P<parameters>.*)`, c.genericCall),
		matcher.WildcardMatcher(c.configTrigger),
	)
}

// e.g. triggered by "trigger job DeployBranch master de3"
func (c *triggerCommand) genericCall(match matcher.Result, message msg.Message) {
	jobName := match.GetString("job")
	if _, ok := c.jobs[jobName]; !ok {
		if len(c.jobs) == 0 {
			c.SendMessage(message, "no job defined in config->jenkins->jobs")
			return
		}

		text := fmt.Sprintf(
			"Sorry, job *%s* is not startable. Possible jobs: \n - *%s*",
			jobName,
			strings.Join(c.cfg.GetSortedNames(), "* \n - *"),
		)
		c.SendMessage(message, text)
		return
	}

	jobConfig := c.jobs[jobName]
	parameterString := strings.TrimSpace(match.GetString("parameters"))

	finalParameters := make(jenkins.Parameters)
	err := jenkins.ParseParameters(jobConfig.config, parameterString, finalParameters)
	if err != nil {
		c.ReplyError(message, err)
		return
	}

	err = jenkins.TriggerJenkinsJob(jobConfig.config, jobName, finalParameters, c.SlackClient, c.jenkins, message)
	if err != nil {
		c.ReplyError(message, err)
		return
	}
}

// check trigger defined in Jenkins.Jobs.*.Trigger
func (c *triggerCommand) configTrigger(ref msg.Ref, text string) bool {
	// start jobs via trigger condition
	for jobName, jobConfig := range c.jobs {
		if jobConfig.trigger == nil {
			continue
		}

		match := jobConfig.trigger.FindStringSubmatch(text)
		if len(match) == 0 {
			continue
		}

		parameters := jobConfig.trigger.ReplaceAllString(text, "")
		jobParams := util.RegexpResultToParams(jobConfig.trigger, match)

		err := jenkins.ParseParameters(jobConfig.config, parameters, jobParams)
		if err != nil {
			c.ReplyError(ref, err)
			return true
		}

		err = jenkins.TriggerJenkinsJob(jobConfig.config, jobName, jobParams, c.SlackClient, c.jenkins, ref.WithText(text))
		if err != nil {
			c.ReplyError(ref, err)
		}

		return true
	}

	return false
}

func (c *triggerCommand) GetHelp() []bot.Help {
	examples := []string{
		"trigger job BuildSomeJob",
		"trigger job DevBackendApiDbCheck master parameter2",
	}

	help := make([]bot.Help, 0, len(c.jobs))
	for jobName, jobConfig := range c.jobs {
		examples = append(examples, fmt.Sprintf("trigger job %s", jobName))

		if jobConfig.config.Trigger == "" {
			continue
		}
		commandHelp := bot.Help{
			Command:     jobName,
			Description: "start jenkins job " + jobName,
			Examples: []string{
				jobConfig.config.Trigger,
			},
			Category: category,
		}
		help = append(help, commandHelp)
	}
	help = append(help, bot.Help{
		Command:     "trigger job <job> [<parameters...>]",
		Description: "start a jenkins build",
		Examples:    examples,
		Category:    category,
	})

	return help
}
