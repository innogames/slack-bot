package pullrequest

import (
	"regexp"
	"strings"
	"text/template"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gitlab.com/gitlab-org/api/client-go"
)

type gitlabFetcher struct {
	client *gitlab.Client
}

func newGitlabCommand(base bot.BaseCommand, cfg *config.Config) bot.Command {
	if cfg.Gitlab.AccessToken == "" && cfg.Gitlab.Host == "" {
		return nil
	}

	options := gitlab.WithBaseURL(cfg.Gitlab.Host)
	gitlabClient, err := gitlab.NewClient(cfg.Gitlab.AccessToken, options)
	if err != nil {
		return nil
	}

	return command{
		base,
		cfg.PullRequest,
		&gitlabFetcher{gitlabClient},
		"(?s).*" + regexp.QuoteMeta(cfg.Gitlab.Host) + "/(?P<repo>.+/.+)/merge_requests/(?P<number>\\d+).*",
	}
}

func (c *gitlabFetcher) getPullRequest(match matcher.Result) (pullRequest, error) {
	var pr pullRequest

	repo := match.GetString("repo")
	repo = strings.TrimSuffix(repo, "/-")

	prNumber := match.GetInt("number")
	rawPullRequest, resp, err := c.client.MergeRequests.GetMergeRequest(
		repo,
		prNumber,
		&gitlab.GetMergeRequestsOptions{},
	)
	if err != nil {
		return pr, err
	}
	resp.Body.Close()

	if rawPullRequest.ID == 0 {
		// aka 404 which is not handles in library anymore
		return pr, errors.New("PR not found")
	}

	return c.convertToPullRequest(rawPullRequest, prNumber), nil
}

func (c *gitlabFetcher) getStatus(pr *gitlab.MergeRequest) prStatus {
	// https://docs.gitlab.com/ce/api/merge_requests.html
	inReview := false
	if len(pr.Reviewers) > 0 {
		inReview = true
	}
	switch {
	case pr.State == "merged":
		return prStatusMerged
	case pr.State == "closed" || pr.State == "locked":
		return prStatusClosed
	case inReview:
		return prStatusInReview
	default:
		return prStatusOpen
	}
}

func (c *gitlabFetcher) getApprovers(rawPullRequest *gitlab.MergeRequest, prNumber int) []string {
	approvers := make([]string, 0)

	state, _, err := c.client.MergeRequestApprovals.GetConfiguration(rawPullRequest.SourceProjectID, prNumber)
	if err != nil {
		log.Errorf("error in gitlab.GetApprovalState: %s", err)
		return approvers
	}

	for _, approver := range state.ApprovedBy {
		approvers = append(approvers, approver.User.Username)
	}

	return approvers
}

func (c *gitlabFetcher) GetTemplateFunction() template.FuncMap {
	return template.FuncMap{
		"gitlabPullRequest": func(repo string, number string) (pullRequest, error) {
			return c.getPullRequest(matcher.Result{
				"repo":   repo,
				"number": number,
			})
		},
	}
}

func (c *gitlabFetcher) getHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "gitlab pull request",
			Category:    category,
			Description: "tracks the state of Gitlab pull requests",
			Examples: []string{
				"https://gitlab.example.com/home-assistant/home-assistant/merge_requests/13958",
			},
		},
	}
}

// convertToPullRequest converts a gitlab.MergeRequest to our own pullRequest structure
func (c *gitlabFetcher) convertToPullRequest(rawPullRequest *gitlab.MergeRequest, prNumber int) pullRequest {
	return pullRequest{
		Name:        rawPullRequest.Title,
		Approvers:   c.getApprovers(rawPullRequest, prNumber),
		Status:      c.getStatus(rawPullRequest),
		Author:      c.getAuthor(rawPullRequest),
		Link:        rawPullRequest.WebURL,
		BuildStatus: c.getPipelineStatus(rawPullRequest),
	}
}

// getPipelineStatus will convert the Pipeline.Status into a buildStatus
// see API: https://docs.gitlab.com/ee/api/pipelines.html
func (c *gitlabFetcher) getPipelineStatus(pr *gitlab.MergeRequest) buildStatus {
	if pr.Pipeline == nil {
		return buildStatusUnknown
	}

	switch pr.Pipeline.Status {
	case "failed":
		return buildStatusFailed
	case "success":
		return buildStatusSuccess
	case "created", "pending", "running":
		return buildStatusRunning
	default:
		return buildStatusUnknown
	}
}

func (c *gitlabFetcher) getAuthor(rawPullRequest *gitlab.MergeRequest) string {
	if rawPullRequest.Author != nil {
		return rawPullRequest.Author.Username
	}
	return ""
}
