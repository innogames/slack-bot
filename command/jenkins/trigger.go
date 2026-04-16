package jenkins

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/innogames/slack-bot/v2/client"
	jenkinsClient "github.com/innogames/slack-bot/v2/command/jenkins/client"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

// command to trigger/start jenkins jobs
type triggerCommand struct {
	jenkinsCommand
	jobs            map[string]triggerCommandData
	cfg             config.JenkinsJobs
	approvals       *approvalStore
	approvalTimeout time.Duration
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
	approvalTimeout time.Duration,
) bot.Command {
	trigger := make(map[string]triggerCommandData, len(jobs))

	for jobName, cfg := range jobs {
		trigger[jobName] = triggerCommandData{
			jobName,
			cfg,
			util.CompileRegexp(cfg.Trigger),
		}
	}

	return &triggerCommand{base, trigger, jobs, newApprovalStore(), approvalTimeout}
}

func (c *triggerCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewRegexpMatcher(`jenkins approve (?P<id>[\w]+)`, c.approveJob),
		matcher.NewRegexpMatcher(`jenkins reject (?P<id>[\w]+)`, c.rejectJob),
		matcher.NewRegexpMatcher(`((trigger|start) (jenkins|build|job)) (?P<job>[\w\-_\\/%.]+)(?P<parameters>.*)`, c.genericCall),
		matcher.WildcardMatcher(c.configTrigger),
	)
}

// e.g. triggered by "trigger job DeployBranch master de3"
func (c *triggerCommand) genericCall(match matcher.Result, message msg.Message) {
	jobName := match.GetString("job")

	// URL decode the job name to handle multibranch pipeline names with encoded characters
	decodedJobName, err := url.QueryUnescape(jobName)
	if err != nil {
		// If decoding fails, use the original job name
		decodedJobName = jobName
	}

	if _, ok := c.jobs[decodedJobName]; !ok {
		if len(c.jobs) == 0 {
			c.SendMessage(message, "no job defined in config: jenkins.jobs")
			return
		}

		text := fmt.Sprintf(
			"Sorry, job *%s* is not startable. Possible jobs: \n - *%s*",
			decodedJobName,
			strings.Join(c.cfg.GetSortedNames(), "* \n - *"),
		)
		c.SendMessage(message, text)
		return
	}

	jobConfig := c.jobs[decodedJobName]
	parameterString := strings.TrimSpace(match.GetString("parameters"))

	finalParameters := make(jenkinsClient.Parameters)
	err = jenkinsClient.ParseParameters(jobConfig.config, parameterString, finalParameters)
	if err != nil {
		c.ReplyError(message, err)
		return
	}

	c.triggerOrRequestApproval(decodedJobName, jobConfig.config, finalParameters, message)
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

		err := jenkinsClient.ParseParameters(jobConfig.config, parameters, jobParams)
		if err != nil {
			c.ReplyError(ref, err)
			return true
		}

		message := ref.WithText(text)
		c.triggerOrRequestApproval(jobName, jobConfig.config, jobParams, message)

		return true
	}

	return false
}

// triggerOrRequestApproval either triggers the job directly or requests approval first
func (c *triggerCommand) triggerOrRequestApproval(jobName string, cfg config.JobConfig, params jenkinsClient.Parameters, message msg.Message) {
	if cfg.NeedsApproval {
		c.requestApproval(jobName, cfg, params, message)
		return
	}

	err := jenkinsClient.TriggerJenkinsJob(cfg, jobName, params, c.SlackClient, c.jenkins, message)
	if err != nil {
		c.ReplyError(message, err)
	}
}

// requestApproval creates a pending approval and sends a DM with approve/reject buttons
func (c *triggerCommand) requestApproval(jobName string, cfg config.JobConfig, params jenkinsClient.Parameters, message msg.Message) {
	id := generateApprovalID()
	now := time.Now()

	approval := &pendingApproval{
		id:        id,
		jobName:   jobName,
		jobConfig: cfg,
		params:    params,
		message:   message,
		createdAt: now,
		expiresAt: now.Add(c.approvalTimeout),
	}
	c.approvals.add(approval)

	// build parameter summary
	var paramText strings.Builder
	for name, value := range params {
		fmt.Fprintf(&paramText, "\n- %s: `%s`", name, value)
	}

	timeoutStr := util.FormatDuration(c.approvalTimeout)

	blocks := []slack.Block{
		client.GetTextBlock(":warning: *Approval Required*"),
		client.GetTextBlock(fmt.Sprintf(
			"Jenkins job *%s* needs your approval before starting.\n\n*Parameters:*%s",
			jobName,
			paramText.String(),
		)),
		client.GetContextBlock(fmt.Sprintf("This approval expires in %s.", timeoutStr)),
		slack.NewActionBlock(
			"",
			client.GetInteractionButton("approve", "Approve", "jenkins approve "+id, slack.StylePrimary),
			client.GetInteractionButton("reject", "Reject", "jenkins reject "+id, slack.StyleDanger),
		),
	}

	c.SlackClient.SendBlockMessageToUser(message.GetUser(), blocks)
	c.SendMessage(message, fmt.Sprintf("Job *%s* requires approval. Please check your direct messages.", jobName))
}

func (c *triggerCommand) approveJob(match matcher.Result, message msg.Message) {
	id := match.GetString("id")
	approval := c.approvals.get(id)

	if approval == nil {
		c.SendMessage(message, "Approval not found or expired. Please re-trigger the job.")
		return
	}

	c.approvals.remove(id)

	log.Infof("Job %s approved by user %s (approval: %s)", approval.jobName, message.GetUser(), id)
	c.SendMessage(message, fmt.Sprintf("Job *%s* approved, starting build...", approval.jobName))

	// trigger the job using the original message context so results go to the original channel
	err := jenkinsClient.TriggerJenkinsJob(approval.jobConfig, approval.jobName, approval.params, c.SlackClient, c.jenkins, approval.message)
	if err != nil {
		c.ReplyError(approval.message, err)
	}
}

func (c *triggerCommand) rejectJob(match matcher.Result, message msg.Message) {
	id := match.GetString("id")
	approval := c.approvals.get(id)

	if approval == nil {
		c.SendMessage(message, "Approval not found or already handled.")
		return
	}

	c.approvals.remove(id)

	log.Infof("Job %s rejected by user %s (approval: %s)", approval.jobName, message.GetUser(), id)
	c.SendMessage(message, fmt.Sprintf("Job *%s* rejected.", approval.jobName))
	c.SendMessage(approval.message, fmt.Sprintf("Job *%s* was rejected.", approval.jobName))
}

// RunAsync periodically cleans up expired approvals
func (c *triggerCommand) RunAsync(ctx *util.ServerContext) {
	ctx.RegisterChild()
	defer ctx.ChildDone()

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.approvals.cleanup()
		case <-ctx.Done():
			return
		}
	}
}

func (c *triggerCommand) GetHelp() []bot.Help {
	examples := []string{
		"trigger job BuildSomeJob",
		"trigger job DevBackendApiDbCheck master parameter2",
	}

	help := make([]bot.Help, 0, len(c.jobs))
	for jobName, jobConfig := range c.jobs {
		examples = append(examples, "trigger job "+jobName)

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
