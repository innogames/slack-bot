package pullrequest

import (
	"context"
	"text/template"

	"github.com/pkg/errors"

	"github.com/google/go-github/github"
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/client"
	"golang.org/x/oauth2"
)

type githubFetcher struct {
	client *github.Client
}

func newGithubCommand(base bot.BaseCommand, cfg *config.Config) bot.Command {
	var githubClient *github.Client
	if cfg.Github.AccessToken == "" {
		githubClient = github.NewClient(client.GetHTTPClient())
	} else {
		ctx := context.Background()

		oauthClient := oauth2.NewClient(
			context.WithValue(ctx, oauth2.HTTPClient, client.GetHTTPClient()),
			oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: cfg.Github.AccessToken},
			),
		)
		githubClient = github.NewClient(oauthClient)
	}

	return command{
		base,
		cfg.PullRequest,
		&githubFetcher{githubClient},
		"(?s).*https://github.com/(?P<project>.+)/(?P<repo>.+)/pull/(?P<number>\\d+).*",
	}
}

func (c *githubFetcher) getPullRequest(match matcher.Result) (pullRequest, error) {
	var pr pullRequest

	project := match.GetString("project")
	repo := match.GetString("repo")
	prNumber := match.GetInt("number")

	ctx := context.Background()
	rawPullRequest, _, err := c.client.PullRequests.Get(ctx, project, repo, prNumber)
	if err != nil {
		var respErr *github.ErrorResponse

		if ok := errors.As(err, &respErr); ok && respErr.Message == "Not Found" {
			return closedPr, nil
		}
		return pr, err
	}

	reviews, _, err := c.client.PullRequests.ListReviews(ctx, project, repo, prNumber, &github.ListOptions{})
	if err != nil {
		return pr, err
	}

	approvers := make([]string, 0)
	inReview := false

	for _, review := range reviews {
		state := review.GetState()
		if state == "COMMENTED" {
			continue
		}
		inReview = true

		if state == "APPROVED" {
			approvers = append(approvers, *review.User.Login)
		}
	}

	pr = pullRequest{
		Name:      rawPullRequest.GetTitle(),
		Status:    c.getStatus(rawPullRequest, inReview),
		Approvers: approvers,
	}

	return pr, nil
}

func (c *githubFetcher) getStatus(pr *github.PullRequest, inReview bool) prStatus {
	switch {
	case pr.GetMerged():
		return prStatusMerged
	case *pr.State == "closed":
		return prStatusClosed
	case inReview:
		return prStatusInReview
	default:
		return prStatusOpen
	}
}

func (c *githubFetcher) GetTemplateFunction() template.FuncMap {
	return template.FuncMap{
		"githubPullRequest": func(project string, repo string, number string) (pullRequest, error) {
			return c.getPullRequest(matcher.Result{
				"project": project,
				"repo":    repo,
				"number":  number,
			})
		},
	}
}

func (c *githubFetcher) getHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "github pull request",
			Description: "tracks the state of Github pull requests",
			Category:    category,
			Examples: []string{
				"https://github.com/home-assistant/home-assistant/pull/13958",
			},
		},
	}
}
