package pullrequest

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
	"regexp"
	"strings"
	"text/template"
)

type gitlabFetcher struct {
	client *gitlab.Client
}

func newGitlabCommand(slackClient client.SlackClient, cfg config.Config, logger *logrus.Logger) bot.Command {
	if cfg.Gitlab.AccessToken == "" && cfg.Gitlab.Host == "" {
		return nil
	}

	options := gitlab.WithBaseURL(cfg.Gitlab.Host)
	gitlabClient, err := gitlab.NewClient(cfg.Gitlab.AccessToken, options)
	if err != nil {
		return nil
	}

	return command{
		cfg.PullRequest,
		slackClient,
		logger,
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

	pr = pullRequest{
		name:      rawPullRequest.Title,
		merged:    rawPullRequest.State == "merged" || rawPullRequest.State == "closed",
		declined:  false,
		approvers: c.getApprovers(rawPullRequest, prNumber),
		inReview:  false,
	}

	return pr, nil
}

func (c *gitlabFetcher) getApprovers(rawPullRequest *gitlab.MergeRequest, prNumber int) []string {
	approvers := make([]string, 0)

	if rawPullRequest.Upvotes > 0 {
		emojis, resp, _ := c.client.AwardEmoji.ListMergeRequestAwardEmoji(
			rawPullRequest.SourceProjectID,
			prNumber,
			&gitlab.ListAwardEmojiOptions{},
		)
		if resp != nil {
			resp.Body.Close()
		}
		for _, emoji := range emojis {
			if emoji.Name == "thumbsup" {
				approvers = append(approvers, emoji.User.Username)
			}
		}
	}

	return approvers
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
			Category:    category,
			Description: "tracks the state of gitlab pull requests",
			Examples: []string{
				"https://gitlab.example.com/home-assistant/home-assistant/merge_requests/13958",
			},
		},
	}
}
