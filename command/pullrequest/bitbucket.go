package pullrequest

import (
	"regexp"
	"slices"
	"strings"
	"text/template"

	bitbucket "github.com/gfleury/go-bitbucket-v1"
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type bitbucketFetcher struct {
	bitbucketClient *bitbucket.DefaultApiService
}

var closedPr = pullRequest{
	Status: prStatusClosed,
}

func newBitbucketCommand(base bot.BaseCommand, cfg *config.Config) bot.Command {
	if !cfg.Bitbucket.IsEnabled() {
		return nil
	}

	bitbucketClient, err := client.GetBitbucketClient(cfg.Bitbucket)
	if err != nil {
		log.Error(errors.Wrap(err, "error while initializing bitbucket client"))
		return nil
	}

	return command{
		base,
		cfg.PullRequest,
		&bitbucketFetcher{bitbucketClient},
		"(?s).*" + regexp.QuoteMeta(cfg.Bitbucket.Host) + "/projects/(?P<project>.+)/repos/(?P<repo>.+)/pull-requests/(?P<number>\\d+).*",
	}
}

func (c *bitbucketFetcher) getPullRequest(match matcher.Result, config *config.PullRequest) (pullRequest, error) {
	var pr pullRequest

	project := match.GetString("project")
	repo := match.GetString("repo")
	number := match.GetInt("number")
	rawResponse, err := c.bitbucketClient.GetPullRequest(project, repo, number)
	if err != nil {
		// handle deleted PR as already "closed" one
		if strings.Contains(err.Error(), "Status: 404") {
			return closedPr, nil
		}

		return pr, errors.Wrap(err, "error while loading data from Bitbucket")
	}

	rawPullRequest, err := bitbucket.GetPullRequestResponse(rawResponse)
	if err != nil {
		return pr, err
	}

	approvers := make([]string, 0)
	for _, reviewer := range rawPullRequest.Reviewers {
		if reviewer.Approved {
			approvers = append(approvers, reviewer.User.Name)
		}
	}

	var author string
	if rawPullRequest.Author != nil {
		author = rawPullRequest.Author.User.Name
	}

	var link string
	if len(rawPullRequest.Links.Self) > 0 {
		link = rawPullRequest.Links.Self[0].Href
	}

	var latestCommentTimestamp int64

	if config.Notifications.NewReviewComments.IsEnabled() {
		newCommentsEnabledForRepo := slices.Contains(config.Notifications.NewReviewComments.Repos, repo)

		if newCommentsEnabledForRepo {
			latestCommentTimestamp, err = c.getLatestReviewCommentTimestamp(&rawPullRequest)
			if err != nil {
				return pr, errors.Wrap(err, "error while loading review comments from Bitbucket")
			}
		}
	}

	pr = pullRequest{
		Name:                          rawPullRequest.Title,
		Status:                        c.getStatus(&rawPullRequest),
		BuildStatus:                   c.getBuildStatus(rawPullRequest.FromRef.LatestCommit),
		Author:                        author,
		Link:                          link,
		Approvers:                     approvers,
		LatestReviewCommentsTimestamp: latestCommentTimestamp,
	}

	return pr, nil
}

func (c *bitbucketFetcher) getStatus(pr *bitbucket.PullRequest) prStatus {
	// https://developer.atlassian.com/bitbucket/api/2/reference/resource/repositories/%7Bworkspace%7D/%7Brepo_slug%7D/pullrequests
	switch pr.State {
	case "MERGED":
		return prStatusMerged
	case "DECLINED", "SUPERSEDED":
		return prStatusClosed
	}

	if len(pr.Reviewers) > 0 {
		return prStatusInReview
	}

	return prStatusOpen
}

// try to extract the current build Status from a PR, based on the recent commit
func (c *bitbucketFetcher) getBuildStatus(lastCommit string) buildStatus {
	status := buildStatusUnknown
	if lastCommit == "" {
		return status
	}

	rawBuilds, err := c.bitbucketClient.GetCommitBuildStatuses(lastCommit)
	if err != nil {
		return status
	}

	builds, err := bitbucket.GetBuildStatusesResponse(rawBuilds)
	if err != nil {
		return status
	}

	for _, build := range builds {
		switch build.State {
		case "SUCCESS", "SUCCESSFUL":
			if status == buildStatusUnknown {
				status = buildStatusSuccess
			}
		case "INPROGRESS":
			status = buildStatusRunning
		case "FAILED":
			if status != buildStatusRunning {
				status = buildStatusFailed
			}
		}
	}

	return status
}

func (c *bitbucketFetcher) getLatestReviewCommentTimestamp(pr *bitbucket.PullRequest) (int64, error) {
	activities, err := c.loadAllPrActivities(pr)
	if err != nil {
		return 0, err
	}

	activities = slices.DeleteFunc(activities, func(activity bitbucket.Activity) bool {
		return activity.Action != bitbucket.ActionCommented
	})

	if len(activities) == 0 {
		return 0, nil
	}

	var latestTimeStamp int64

	for _, activity := range activities {
		reviewComments := getAllReviewComments(&activity.Comment, pr)
		for _, reviewComment := range reviewComments {
			latestTimeStamp = max(latestTimeStamp, reviewComment.CreatedDate)
		}
	}

	return latestTimeStamp, nil
}

func getAllReviewComments(comment *bitbucket.ActivityComment, pr *bitbucket.PullRequest) []bitbucket.ActivityComment {
	if comment == nil {
		return nil
	}

	var reviewComments []bitbucket.ActivityComment

	if isCommentFromRealReviewer(comment, pr) {
		reviewComments = append(reviewComments, *comment)
	}

	for _, nestedComment := range comment.Comments {
		nestedReviewComments := getAllReviewComments(&nestedComment, pr)
		reviewComments = append(reviewComments, nestedReviewComments...)
	}

	return reviewComments
}

func isCommentFromRealReviewer(comment *bitbucket.ActivityComment, pr *bitbucket.PullRequest) bool {
	if len(pr.Reviewers) == 0 {
		return false
	}

	isReviewer := slices.ContainsFunc(pr.Reviewers, func(userMetadata bitbucket.UserWithMetadata) bool {
		return userMetadata.User.Name == comment.Author.Name
	})

	return isReviewer
}

func (c *bitbucketFetcher) loadAllPrActivities(pr *bitbucket.PullRequest) ([]bitbucket.Activity, error) {
	apiResponse, err := c.bitbucketClient.GetActivities(pr.FromRef.Repository.Project.Key, pr.FromRef.Repository.Slug, pr.ID, nil)
	if err != nil {
		return nil, err
	}

	activities, err := bitbucket.GetActivitiesResponse(apiResponse)
	if err != nil {
		return nil, err
	}

	allActivities := activities.Values

	hasNextPage, nextPage := bitbucket.HasNextPage(apiResponse)
	requestOptions := map[string]any{}

	for hasNextPage {
		requestOptions["start"] = nextPage
		apiResponse, err := c.bitbucketClient.GetActivities(pr.FromRef.Repository.Project.Key, pr.FromRef.Repository.Slug, pr.ID, requestOptions)
		if err != nil {
			return allActivities, err
		}

		activities, err := bitbucket.GetActivitiesResponse(apiResponse)
		if err != nil {
			return allActivities, err
		}

		allActivities = append(allActivities, activities.Values...)
		hasNextPage, nextPage = bitbucket.HasNextPage(apiResponse)
	}

	return allActivities, nil
}

func (c *bitbucketFetcher) GetTemplateFunction(cfg *config.PullRequest) template.FuncMap {
	return template.FuncMap{
		"bitbucketPullRequest": func(project string, repo string, number string) (pullRequest, error) {
			return c.getPullRequest(matcher.Result{
				"project": project,
				"repo":    repo,
				"number":  number,
			}, cfg)
		},
	}
}

func (c *bitbucketFetcher) getHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "bitbucket pull request",
			Category:    category,
			Description: "tracks the state of bitbucket/stash pull requests",
			Examples: []string{
				"https://bitbucket.example.com/home-assistant/home-assistant/pull-requests/13958",
			},
		},
	}
}
