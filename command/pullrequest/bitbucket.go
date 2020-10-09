package pullrequest

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/xoom/stash"
	"net/url"
	"regexp"
	"text/template"
)

type bitbucketFetcher struct {
	bitbucketClient stash.Stash
}

func newBitbucketCommand(slackClient client.SlackClient, cfg config.Config) bot.Command {
	if !cfg.Bitbucket.IsEnabled() {
		return nil
	}

	host, _ := url.Parse(cfg.Bitbucket.Host)
	bitbucketClient := stash.NewClient(cfg.Bitbucket.Username, cfg.Bitbucket.Password, host)

	return &command{
		cfg.PullRequest,
		slackClient,
		&bitbucketFetcher{bitbucketClient},
		"(?s).*" + regexp.QuoteMeta(cfg.Bitbucket.Host) + "/projects/(?P<project>.+)/repos/(?P<repo>.+)/pull-requests/(?P<number>\\d+).*",
	}
}

func (c *bitbucketFetcher) getPullRequest(match matcher.Result) (pullRequest, error) {
	var pr pullRequest

	project := match.GetString("project")
	repo := match.GetString("repo")
	number := match.GetString("number")
	rawPullRequest, err := c.bitbucketClient.GetPullRequest(project, repo, number)
	if err != nil {
		return pr, err
	}

	fmt.Println(rawPullRequest)
	fmt.Println(rawPullRequest.State)

	approvers := make([]string, 0)
	for _, reviewer := range rawPullRequest.Reviewers {
		if reviewer.Approved {
			approvers = append(approvers, reviewer.User.Name)
		}
	}

	pr = pullRequest{
		name:      rawPullRequest.Title,
		merged:    rawPullRequest.State == "MERGED",
		declined:  rawPullRequest.State == "DECLINED",
		approvers: approvers,
		inReview:  len(rawPullRequest.Reviewers) > 0,
	}

	return pr, nil
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
			Description: "tracks the state of bitbucket/stash pull requests",
			Examples: []string{
				"https://bitbucket.example.com/home-assistant/home-assistant/pull-requests/13958",
			},
		},
	}
}
