package pullrequest

import (
	bitbucket "github.com/gfleury/go-bitbucket-v1"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"regexp"
	"text/template"
)

type bitbucketFetcher struct {
	bitbucketClient *bitbucket.DefaultApiService
}

func newBitbucketCommand(base bot.BaseCommand, cfg config.Config) bot.Command {
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

func (c *bitbucketFetcher) getPullRequest(match matcher.Result) (pullRequest, error) {
	var pr pullRequest

	project := match.GetString("project")
	repo := match.GetString("repo")
	number := match.GetInt("number")
	rawResponse, err := c.bitbucketClient.GetPullRequest(project, repo, number)
	if err != nil {
		return pr, err
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

	pr = pullRequest{
		Name:        rawPullRequest.Title,
		merged:      rawPullRequest.State == "MERGED",
		declined:    rawPullRequest.State == "DECLINED",
		approvers:   approvers,
		inReview:    len(rawPullRequest.Reviewers) > 0,
		buildStatus: c.getBuildStatus(rawPullRequest.FromRef.LatestCommit),
	}

	return pr, nil
}

// try to extract the current build status from a PR, based on the recent commit
func (c *bitbucketFetcher) getBuildStatus(lastCommit string) buildStatus {
	buildStatus := buildStatusUnknown
	if lastCommit == "" {
		return buildStatus
	}

	rawBuilds, err := c.bitbucketClient.GetCommitBuildStatuses(lastCommit)
	if err != nil {
		return buildStatus
	}

	builds, err := bitbucket.GetBuildStatusesResponse(rawBuilds)
	if err != nil {
		return buildStatus
	}
	for _, build := range builds {
		switch build.State {
		case "SUCCESS":
			return buildStatusSuccess
		case "INPROGRESS":
			return buildStatusRunning
		case "FAILED":
			return buildStatusFailed
		}
	}

	return buildStatus
}

func (c *bitbucketFetcher) GetTemplateFunction() template.FuncMap {
	return template.FuncMap{
		"bitbucketPullRequest": func(project string, repo string, number string) (pullRequest, error) {
			return c.getPullRequest(matcher.MapResult{
				"project": project,
				"repo":    repo,
				"number":  number,
			})
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
