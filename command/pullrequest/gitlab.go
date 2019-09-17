package pullrequest

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/xanzy/go-gitlab"
	"net/http"
	"regexp"
	"text/template"
)

type gitlabFetcher struct {
	client *gitlab.Client
}

func newGitlabCommand(slackClient client.SlackClient, cfg config.Config) bot.Command {
	if cfg.Gitlab.AccessToken == "" && cfg.Gitlab.Host == "" {
		return nil
	}

	client := gitlab.NewClient(&http.Client{}, cfg.Gitlab.AccessToken)
	err := client.SetBaseURL(cfg.Gitlab.Host + "/api/v4")
	if err != nil {
		return nil
	}

	return &command{
		slackClient,
		&gitlabFetcher{client},
		"(?s).*" + regexp.QuoteMeta(cfg.Gitlab.Host) + "/(?P<repo>.+/.+)/merge_requests/(?P<number>\\d+).*",
	}
}

func (c *gitlabFetcher) getPullRequest(match matcher.Result) (pullRequest, error) {
	var pr pullRequest

	prNumber := match.GetInt("number")
	rawPullRequest, _, err := c.client.MergeRequests.GetMergeRequest(
		match.GetString("repo"),
		prNumber,
		&gitlab.GetMergeRequestsOptions{},
	)
	if err != nil {
		return pr, err
	}

	pr = pullRequest{
		name:     rawPullRequest.Title,
		merged:   rawPullRequest.State == "merged" || rawPullRequest.State == "closed",
		declined: false,
		approved: rawPullRequest.Upvotes > 0,
		inReview: false,
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
			"gitlab pull request",
			"tracks the state of gitlab pull requests",
			[]string{
				"https://gitlab.example.com/home-assistant/home-assistant/merge_requests/13958",
			},
		},
	}
}
