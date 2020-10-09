package pullrequest

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/xanzy/go-gitlab"
	"regexp"
	"strings"
	"text/template"
)

type gitlabFetcher struct {
	client *gitlab.Client
}

func newGitlabCommand(slackClient client.SlackClient, cfg config.Config) bot.Command {
	if cfg.Gitlab.AccessToken == "" && cfg.Gitlab.Host == "" {
		return nil
	}

	options := gitlab.WithBaseURL(cfg.Gitlab.Host)
	client, err := gitlab.NewClient(cfg.Gitlab.AccessToken, options)
	if err != nil {
		return nil
	}

	return &command{
		cfg.PullRequest,
		slackClient,
		&gitlabFetcher{client},
		"(?s).*" + regexp.QuoteMeta(cfg.Gitlab.Host) + "/(?P<repo>.+/.+)/merge_requests/(?P<number>\\d+).*",
	}
}

func (c *gitlabFetcher) getPullRequest(match matcher.Result) (pullRequest, error) {
	var pr pullRequest

	repo := match.GetString("repo")
	repo = strings.TrimSuffix(repo, "/-")

	prNumber := match.GetInt("number")
	rawPullRequest, _, err := c.client.MergeRequests.GetMergeRequest(
		repo,
		prNumber,
		&gitlab.GetMergeRequestsOptions{},
	)
	if err != nil {
		return pr, err
	}

	approvers := make([]string, 0)
	if rawPullRequest.Upvotes > 0 {
		approvers = append(approvers, "unknown")
	}

	pr = pullRequest{
		name:      rawPullRequest.Title,
		merged:    rawPullRequest.State == "merged" || rawPullRequest.State == "closed",
		declined:  false,
		approvers: approvers,
		inReview:  false,
	}

	return pr, nil
}

func (c *gitlabFetcher) GetTemplateFunction() template.FuncMap {
	return template.FuncMap{
		"gitlabPullRequest": func(repo string, number string) (pullRequest, error) {
			return c.getPullRequest(matcher.MapResult{
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
			Description: "tracks the state of gitlab pull requests",
			Examples: []string{
				"https://gitlab.example.com/home-assistant/home-assistant/merge_requests/13958",
			},
		},
	}
}
