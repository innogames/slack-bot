package gitlab

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/innogames/slack-bot/v2/command/queue"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

const (
	iconRunning = "arrows_counterclockwise"
	iconSuccess = "white_check_mark"
	iconFailed  = "x"

	colorRunning = "#E0E000"
	colorSuccess = "#00EE00"
	colorFailed  = "#CC0000"
	colorOther   = "#CCCCCC"

	pollInterval = 15 * time.Second
)

type urlType int

const (
	urlTypePipeline urlType = iota
	urlTypeJob
)

func (t urlType) String() string {
	if t == urlTypeJob {
		return "job"
	}
	return "pipeline"
}

type parsedURL struct {
	project string
	kind    urlType
	id      int
}

// gitlabAPI abstracts the GitLab API calls for testability
type gitlabAPI interface {
	GetPipeline(pid any, pipeline int64) (*gitlab.Pipeline, error)
	ListPipelineJobs(pid any, pipeline int64) ([]*gitlab.Job, error)
	GetJob(pid any, jobID int64) (*gitlab.Job, error)
}

type realGitlabAPI struct {
	client *gitlab.Client
}

func (g *realGitlabAPI) GetPipeline(pid any, pipeline int64) (*gitlab.Pipeline, error) {
	p, resp, err := g.client.Pipelines.GetPipeline(pid, pipeline)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	return p, err
}

func (g *realGitlabAPI) ListPipelineJobs(pid any, pipeline int64) ([]*gitlab.Job, error) {
	opts := &gitlab.ListJobsOptions{
		ListOptions: gitlab.ListOptions{PerPage: 100},
	}
	jobs, resp, err := g.client.Jobs.ListPipelineJobs(pid, pipeline, opts)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	return jobs, err
}

func (g *realGitlabAPI) GetJob(pid any, jobID int64) (*gitlab.Job, error) {
	j, resp, err := g.client.Jobs.GetJob(pid, jobID)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	return j, err
}

type gitlabCommand struct {
	bot.BaseCommand
	api  gitlabAPI
	host string
}

type notifyCommand struct {
	gitlabCommand
}

func newNotifyCommand(base gitlabCommand) bot.Command {
	return &notifyCommand{base}
}

func (c *notifyCommand) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher(`gitlab notify (?P<url>https://\S+)`, c.watch)
}

func (c *notifyCommand) watch(match matcher.Result, message msg.Message) {
	rawURL := match.GetString("url")

	// Validate host
	parsedURLObj, err := url.Parse(rawURL)
	if err != nil {
		c.SendMessage(message, "Invalid URL")
		return
	}
	configuredHost, _ := url.Parse(c.host)
	if parsedURLObj.Host != configuredHost.Host {
		c.SendMessage(message, "URL does not match configured GitLab host")
		return
	}

	parsed, err := parseGitlabURL(rawURL)
	if err != nil {
		c.SendMessage(message, fmt.Sprintf("Invalid GitLab URL: %s", err))
		return
	}

	status, err := c.fetchStatus(parsed)
	if err != nil {
		c.SendMessage(message, fmt.Sprintf("Error fetching GitLab %s: %s", parsed.kind, err))
		return
	}

	if isTerminalStatus(status.state) {
		c.SendMessage(message, fmt.Sprintf("GitLab %s *%s/%d* already finished with status: *%s*", parsed.kind, parsed.project, parsed.id, status.state))
		return
	}

	title := fmt.Sprintf("Watching GitLab %s %s/%d", parsed.kind, parsed.project, parsed.id)
	attachment := buildAttachment(title, status, rawURL)
	msgTimestamp := c.SendMessage(message, "", slack.MsgOptionAttachments(attachment))

	runningCommand := queue.AddRunningCommand(
		message,
		"gitlab notify "+rawURL,
	)

	c.AddReaction(iconRunning, message)

	go c.pollUntilDone(message, parsed, rawURL, title, msgTimestamp, runningCommand)
}

type jobDetail struct {
	name   string
	stage  string
	status string
	webURL string
}

type statusInfo struct {
	state      string
	summary    string
	jobDetails []jobDetail // running and failed jobs
	jobsDone   int         // number of finished jobs (for progress bar)
	jobsTotal  int         // total number of jobs
	duration   time.Duration
	ref        string
}

func (c *notifyCommand) fetchStatus(parsed parsedURL) (statusInfo, error) {
	switch parsed.kind {
	case urlTypePipeline:
		pipeline, err := c.api.GetPipeline(parsed.project, int64(parsed.id))
		if err != nil {
			return statusInfo{}, err
		}

		info := statusInfo{
			state:    pipeline.Status,
			duration: time.Duration(pipeline.Duration) * time.Second,
			ref:      pipeline.Ref,
		}

		jobs, err := c.api.ListPipelineJobs(parsed.project, int64(parsed.id))
		if err != nil {
			return info, err
		}
		info.summary = buildJobSummary(jobs)
		info.jobDetails = collectNotableJobs(jobs)
		info.jobsTotal = len(jobs)
		info.jobsDone = countDoneJobs(jobs)

		return info, nil

	case urlTypeJob:
		job, err := c.api.GetJob(parsed.project, int64(parsed.id))
		if err != nil {
			return statusInfo{}, err
		}

		return statusInfo{
			state:    job.Status,
			summary:  fmt.Sprintf("Job: %s (stage: %s)", job.Name, job.Stage),
			duration: time.Duration(job.Duration) * time.Second,
			ref:      job.Ref,
		}, nil
	}

	return statusInfo{}, errors.New("unknown URL type")
}

func (c *notifyCommand) pollUntilDone(
	message msg.Message,
	parsed parsedURL,
	rawURL string,
	title string,
	msgTimestamp string,
	runningCommand *queue.RunningCommand,
) {
	defer runningCommand.Done()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	var status statusInfo
	for range ticker.C {
		var err error
		status, err = c.fetchStatus(parsed)
		if err != nil {
			log.Warnf("Error polling GitLab %s %s#%d: %s", parsed.kind, parsed.project, parsed.id, err)
			continue
		}

		attachment := buildAttachment(title, status, rawURL)
		c.SendMessage(
			message,
			"",
			slack.MsgOptionUpdate(msgTimestamp),
			slack.MsgOptionAttachments(attachment),
		)

		if isTerminalStatus(status.state) {
			break
		}
	}

	c.RemoveReaction(iconRunning, message)
	if status.state == "success" {
		c.AddReaction(iconSuccess, message)
	} else {
		c.AddReaction(iconFailed, message)
	}

	c.SendMessage(message, fmt.Sprintf(
		"<@%s> GitLab %s *%s*: %s/%d %s in %s",
		message.User,
		parsed.kind,
		status.state,
		parsed.project,
		parsed.id,
		rawURL,
		util.FormatDuration(status.duration),
	))
}

func buildAttachment(title string, status statusInfo, webURL string) slack.Attachment {
	color := statusColor(status.state)

	attachment := slack.Attachment{
		Title:     title,
		TitleLink: webURL,
		Color:     color,
	}

	attachment.Fields = append(attachment.Fields, slack.AttachmentField{
		Title: "Status",
		Value: status.state,
		Short: true,
	})

	if status.ref != "" {
		attachment.Fields = append(attachment.Fields, slack.AttachmentField{
			Title: "Ref",
			Value: status.ref,
			Short: true,
		})
	}

	if progress := util.RenderCountProgressBar(status.jobsDone, status.jobsTotal); progress != "" {
		attachment.Fields = append(attachment.Fields, slack.AttachmentField{
			Title: "Progress",
			Value: progress,
			Short: false,
		})
	}

	if status.summary != "" {
		attachment.Fields = append(attachment.Fields, slack.AttachmentField{
			Title: "Jobs",
			Value: status.summary,
			Short: false,
		})
	}

	if details := formatJobDetails(status.jobDetails); details != "" {
		attachment.Fields = append(attachment.Fields, slack.AttachmentField{
			Title: "Running/Failed Jobs",
			Value: details,
			Short: false,
		})
	}

	attachment.Actions = []slack.AttachmentAction{
		client.GetSlackLink("View in GitLab", webURL),
	}

	return attachment
}

func statusColor(status string) string {
	switch status {
	case "success":
		return colorSuccess
	case "failed":
		return colorFailed
	case "running", "pending", "created", "waiting_for_resource", "preparing":
		return colorRunning
	default:
		return colorOther
	}
}

var terminalStatuses = map[string]bool{
	"success":  true,
	"failed":   true,
	"canceled": true,
	"skipped":  true,
	"manual":   true,
}

func isTerminalStatus(status string) bool {
	return terminalStatuses[status]
}

// buildJobSummary counts jobs by status and returns a human-readable summary
func buildJobSummary(jobs []*gitlab.Job) string {
	counts := make(map[string]int)
	for _, job := range jobs {
		counts[job.Status]++
	}

	order := []string{"running", "success", "failed", "pending", "created", "canceled", "skipped", "manual"}
	var parts []string
	for _, s := range order {
		if n := counts[s]; n > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", n, s))
		}
	}

	return strings.Join(parts, ", ")
}

// countDoneJobs returns the number of jobs in a terminal state
func countDoneJobs(jobs []*gitlab.Job) int {
	count := 0
	for _, job := range jobs {
		if isTerminalStatus(job.Status) {
			count++
		}
	}
	return count
}

// collectNotableJobs returns running and failed jobs for detailed display
func collectNotableJobs(jobs []*gitlab.Job) []jobDetail {
	var details []jobDetail
	for _, job := range jobs {
		if job.Status == "running" || job.Status == "failed" {
			details = append(details, jobDetail{
				name:   job.Name,
				stage:  job.Stage,
				status: job.Status,
				webURL: job.WebURL,
			})
		}
	}
	return details
}

// formatJobDetails renders running/failed jobs as a Slack mrkdwn list with links
func formatJobDetails(details []jobDetail) string {
	if len(details) == 0 {
		return ""
	}

	var lines []string
	for _, d := range details {
		icon := ":arrow_forward:"
		if d.status == "failed" {
			icon = ":x:"
		}
		lines = append(lines, fmt.Sprintf("%s <%s|%s> (%s)", icon, d.webURL, d.name, d.stage))
	}
	return strings.Join(lines, "\n")
}

// parseGitlabURL extracts the project path, resource type, and ID from a GitLab URL
// Supports: /<project>/-/pipelines/<id> and /<project>/-/jobs/<id>
func parseGitlabURL(rawURL string) (parsedURL, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return parsedURL{}, err
	}

	path := strings.TrimSuffix(u.Path, "/")

	types := []struct {
		segment string
		kind    urlType
	}{
		{"/-/pipelines/", urlTypePipeline},
		{"/-/jobs/", urlTypeJob},
	}

	for _, t := range types {
		idx := strings.Index(path, t.segment)
		if idx < 0 {
			continue
		}
		project := strings.TrimPrefix(path[:idx], "/")
		idStr := path[idx+len(t.segment):]
		// strip any trailing path after the ID
		if slashIdx := strings.Index(idStr, "/"); slashIdx >= 0 {
			idStr = idStr[:slashIdx]
		}
		id, parseErr := strconv.Atoi(idStr)
		if parseErr != nil {
			return parsedURL{}, fmt.Errorf("invalid %s ID: %s", t.kind, idStr)
		}
		return parsedURL{
			project: project,
			kind:    t.kind,
			id:      id,
		}, nil
	}

	return parsedURL{}, errors.New("URL must contain /-/pipelines/<id> or /-/jobs/<id>")
}

func (c *notifyCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "gitlab notify <url>",
			Description: "watch a GitLab pipeline or job and notify when it finishes",
			Examples: []string{
				"gitlab notify https://gitlab.example.com/my-group/my-project/-/pipelines/12345",
				"gitlab notify https://gitlab.example.com/my-group/my-project/-/jobs/67890",
			},
			Category: category,
		},
	}
}
