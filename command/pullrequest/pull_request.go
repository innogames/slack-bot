package pullrequest

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/command/queue"
	"github.com/nlopes/slack"
	"time"
)

const (
	iconInReview    = "eyes"
	iconApproved    = "white_check_mark"
	iconMerged      = "twisted_rightwards_arrows"
	iconDeclined    = "x"
	iconBuildFailed = "red_circle"
	iconError       = "x"
	checkInterval   = time.Second * 20
)

type buildStatus int

const (
	buildStatusUnknown buildStatus = iota
	buildStatusSucccess
	buildStatusFailed
	buildStatusRunning
)

type fetcher interface {
	getPullRequest(match matcher.Result) (pullRequest, error)
	getHelp() []bot.Help
}

type command struct {
	slackClient client.SlackClient
	fetcher     fetcher
	regexp      string
}

type pullRequest struct {
	name        string
	declined    bool
	merged      bool
	approved    bool
	inReview    bool
	buildStatus buildStatus
}

func (c *command) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher(c.regexp, c.Execute)
}

func (c *command) Execute(match matcher.Result, event slack.MessageEvent) {
	_, err := c.fetcher.getPullRequest(match)

	if err != nil {
		c.slackClient.ReplyError(event, err)
		return
	}

	go c.watch(match, event)
}

func (c *command) watch(match matcher.Result, event slack.MessageEvent) {
	msgRef := slack.NewRefToMessage(event.Channel, event.Timestamp)

	inReview := false
	hasApproval := false
	failedBuild := false
	done := queue.AddRunningCommand(event, event.Text)
	defer func() {
		done <- true
	}()

	for {
		pr, err := c.fetcher.getPullRequest(match)
		if err != nil {
			// reply error in new thread
			c.slackClient.SendMessage(event, err.Error(), slack.MsgOptionTS(event.Timestamp))
			c.slackClient.AddReaction(iconError, msgRef)
			return
		}

		if pr.merged {
			// PR got merged
			c.slackClient.RemoveReaction(iconInReview, msgRef)
			c.slackClient.AddReaction(iconMerged, msgRef)

			return
		}

		if pr.declined {
			// PR got declined
			c.slackClient.RemoveReaction(iconInReview, msgRef)
			c.slackClient.RemoveReaction(iconApproved, msgRef)
			c.slackClient.AddReaction(iconDeclined, msgRef)

			return
		}

		if pr.approved && !hasApproval {
			c.slackClient.RemoveReaction(iconInReview, msgRef)
			c.slackClient.RemoveReaction(iconDeclined, msgRef)
			c.slackClient.AddReaction(iconApproved, msgRef)
			hasApproval = true
		}

		if pr.inReview {
			if !hasApproval && !inReview {
				c.slackClient.AddReaction(iconInReview, msgRef)
				inReview = true
			}
		}

		// monitor build status
		if pr.buildStatus == buildStatusFailed && !failedBuild {
			c.slackClient.AddReaction(iconBuildFailed, msgRef)
			failedBuild = true
		} else if pr.buildStatus != buildStatusFailed && failedBuild {
			c.slackClient.RemoveReaction(iconBuildFailed, msgRef)
			failedBuild = false
		}

		time.Sleep(checkInterval)
	}
}

func (c *command) GetHelp() []bot.Help {
	return c.fetcher.getHelp()
}
