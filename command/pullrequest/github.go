package pullrequest

import (
	"context"
	"github.com/google/go-github/github"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"text/template"
)

type githubFetcher struct {
	client *github.Client
}

func newGithubCommand(slackClient client.SlackClient, cfg config.Config, logger *logrus.Logger) bot.Command {
	if cfg.Github.AccessToken == "" {
		return nil
	}

	ctx := context.Background()
	oauthClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.Github.AccessToken},
	))
	githubClient := github.NewClient(oauthClient)

	return command{
		cfg.PullRequest,
		slackClient,
		logger,
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
	rawPullRequest, resp, err := c.client.PullRequests.Get(ctx, project, repo, prNumber)
	if err != nil {
		return pr, err
	}
	resp.Body.Close()

	reviews, resp, err := c.client.PullRequests.ListReviews(ctx, project, repo, prNumber, &github.ListOptions{})
	if err != nil {
		return pr, err
	}
	resp.Body.Close()

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
		name:      rawPullRequest.GetTitle(),
		merged:    rawPullRequest.GetMerged(),
		closed:    *rawPullRequest.State == "closed",
		declined:  false,
		approvers: approvers,
		inReview:  inReview,
	}

	return pr, nil
}

func (c *githubFetcher) GetTemplateFunction() template.FuncMap {
	return template.FuncMap{
		"githubPullRequest": func(project string, repo string, number string) (pullRequest, error) {
			return c.getPullRequest(matcher.MapResult{
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
			Description: "tracks the state of github pull requests",
			Examples: []string{
				"https://github.com/home-assistant/home-assistant/pull/13958",
			},
		},
	}
}
