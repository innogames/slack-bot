package jenkins

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/client/jenkins"
	"github.com/nlopes/slack"
	"github.com/sirupsen/logrus"
	"regexp"
	"sort"
	"strings"
)

type triggerCommand struct {
	jenkins     jenkins.Client
	slackClient client.SlackClient
	jobs        map[string]triggerCommandData
	logger      *logrus.Logger
}

type triggerCommandData struct {
	jobName string
	config  config.JobConfig
	trigger *regexp.Regexp
}

// newTriggerCommand initialize a new command to trigger for whitelisted jenkins job
func newTriggerCommand(
	jenkins jenkins.Client,
	slackClient client.SlackClient,
	jobs config.JenkinsJobs,
	logger *logrus.Logger,
) bot.Command {
	trigger := make(map[string]triggerCommandData, len(jobs))

	for jobName, cfg := range jobs {
		trigger[jobName] = triggerCommandData{
			jobName,
			cfg,
			util.CompileRegexp(cfg.Trigger),
		}
	}

	return &triggerCommand{jenkins, slackClient, trigger, logger}
}

func (c *triggerCommand) IsEnabled() bool {
	return c.jenkins != nil
}

func (c *triggerCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewRegexpMatcher("((trigger|start) (jenkins|build|job)) (?P<job>[\\w\\-_]*)(?P<parameters>.*)", c.GenericCall),
		matcher.NewConditionalMatcher(c.ConfigTrigger),
	)
}

// e.g. triggered by "trigger job DeployBranch master de3"
func (c *triggerCommand) GenericCall(match matcher.Result, event slack.MessageEvent) {
	jobName := match.GetString("job")
	if _, ok := c.jobs[jobName]; !ok {
		if len(c.jobs) == 0 {
			c.slackClient.Reply(event, "no job defined in config->jira->jobs")
			return
		}
		message := fmt.Sprintf(
			"Sorry, job *%s* is not startable. Possible jobs: \n - *%s*",
			jobName,
			strings.Join(c.getAllowedJobNames(), "* \n - *"),
		)
		c.slackClient.Reply(event, message)
		return
	}

	jobConfig := c.jobs[jobName]
	parameterString := strings.TrimSpace(match.GetString("parameters"))

	finalParameters := make(jenkins.Parameters)
	err := jenkins.ParseParameters(jobConfig.config, parameterString, finalParameters)
	if err != nil {
		c.slackClient.ReplyError(event, err)
		return
	}

	err = jenkins.TriggerJenkinsJob(jobConfig.config, jobName, finalParameters, c.slackClient, c.jenkins, event, c.logger)
	if err != nil {
		c.slackClient.ReplyError(event, err)
		return
	}
}

// check trigger defined in Jenkins.Jobs.*.Trigger
func (c *triggerCommand) ConfigTrigger(event slack.MessageEvent) bool {
	// start jobs via trigger condition
	for jobName, jobConfig := range c.jobs {
		if jobConfig.trigger == nil {
			continue
		}

		match := jobConfig.trigger.FindStringSubmatch(event.Text)
		if len(match) == 0 {
			continue
		}

		parameters := jobConfig.trigger.ReplaceAllString(event.Text, "")
		jobParams := util.RegexpResultToParams(jobConfig.trigger, match)

		err := jenkins.ParseParameters(jobConfig.config, parameters, jobParams)
		if err != nil {
			c.slackClient.ReplyError(event, err)
			return true
		}

		err = jenkins.TriggerJenkinsJob(jobConfig.config, jobName, jobParams, c.slackClient, c.jenkins, event, c.logger)
		if err != nil {
			c.slackClient.ReplyError(event, err)
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

	var help []bot.Help

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
		}
		help = append(help, commandHelp)
	}
	help = append(help, bot.Help{
		Command:     "jenkins",
		Description: "start a jenkins build",
		Examples:    examples,
	})

	return help
}

func (c *triggerCommand) getAllowedJobNames() []string {
	jobNames := make([]string, 0, len(c.jobs))
	for jobName := range c.jobs {
		jobNames = append(jobNames, jobName)
	}
	sort.Strings(jobNames)

	return jobNames
}
