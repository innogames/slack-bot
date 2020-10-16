package pullrequest

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/command/queue"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"net"
	"time"
)

const (
	iconInReview        = "eyes"
	iconApproved        = "white_check_mark"
	iconMerged          = "twisted_rightwards_arrows"
	iconDeclined        = "x"
	iconBuildFailed     = "red_circle"
	iconError           = "x"
	checkInterval       = time.Second * 30
	maxConnectionErrors = 5
)

type buildStatus int

const (
	buildStatusUnknown buildStatus = iota
	buildStatusSuccess
	buildStatusFailed
	buildStatusRunning
)

type fetcher interface {
	getPullRequest(match matcher.Result) (pullRequest, error)
	getHelp() []bot.Help
}

type command struct {
	cfg         config.PullRequest
	slackClient client.SlackClient
	logger      *logrus.Logger
	fetcher     fetcher
	regexp      string
}

type pullRequest struct {
	name        string
	declined    bool
	merged      bool
	closed      bool
	approvers   []string
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

	hasApproval := false
	failedBuild := false
	connectionErrors := 0

	done := queue.AddRunningCommand(event, event.Text)
	defer func() {
		done <- true
	}()

	for {
		pr, err := c.fetcher.getPullRequest(match)
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
				time.Sleep(checkInterval)
				continue
			}

			connectionErrors++
			if connectionErrors > maxConnectionErrors {
				// reply error in new thread
				c.slackClient.ReplyError(
					event,
					errors.Wrapf(err, "Error while fetching PR data %d times in a row", connectionErrors),
				)
				c.slackClient.AddReaction(iconError, msgRef)
				return
			}
			continue
		}

		connectionErrors = 0
		currentReactions := c.getReactions(err, msgRef)

		// add approved reaction(s)
		if len(pr.approvers) > 0 {
			for icon := range c.getApproveIcons(pr.approvers) {
				c.addReaction(currentReactions, icon, msgRef)
			}

			hasApproval = true
		}

		// add :eyes: when someone is reviewing the PR but nobody approved it yet
		if pr.inReview && !hasApproval && !pr.merged {
			c.addReaction(currentReactions, iconInReview, msgRef)
		} else {
			c.removeReaction(currentReactions, iconInReview, msgRef)
		}

		// add merged reaction
		if pr.merged || pr.closed {
			c.addReaction(currentReactions, iconMerged, msgRef)

			return
		}

		// add declined reaction
		if pr.declined {
			c.removeReaction(currentReactions, iconApproved, msgRef)
			c.addReaction(currentReactions, iconDeclined, msgRef)

			return
		}

		// monitor build status
		if pr.buildStatus == buildStatusFailed && !failedBuild {
			c.addReaction(currentReactions, iconBuildFailed, msgRef)
			failedBuild = true
		} else if pr.buildStatus != buildStatusFailed && failedBuild {
			c.removeReaction(currentReactions, iconBuildFailed, msgRef)
			failedBuild = false
		}

		time.Sleep(checkInterval)
	}
}

// get the current reactions in the given message
func (c *command) getReactions(err error, msgRef slack.ItemRef) map[string]bool {
	currentReactions := make(map[string]bool)
	reactions, err := c.slackClient.GetReactions(msgRef, slack.NewGetReactionsParameters())
	for _, reaction := range reactions {
		currentReactions[reaction.Name] = true
	}

	return currentReactions
}

func (c *command) removeReaction(currentReactions map[string]bool, icon string, msgRef slack.ItemRef) {
	if ok := currentReactions[icon]; !ok {
		// already removed
		return
	}

	delete(currentReactions, icon)
	c.slackClient.RemoveReaction(icon, msgRef)
}

func (c *command) addReaction(currentReactions map[string]bool, icon string, msgRef slack.ItemRef) {
	if _, ok := currentReactions[icon]; ok {
		// already added
		return
	}

	currentReactions[icon] = true
	c.slackClient.AddReaction(icon, msgRef)
}

// generates a map of all icons for the given approvers list. If there is no special mapping, it returns the default icon
func (c *command) getApproveIcons(approvers []string) map[string]bool {
	icons := make(map[string]bool)

	for _, approver := range approvers {
		if icon, ok := c.cfg.CustomApproveReaction[approver]; ok {
			icons[icon] = true
		} else {
			c.logger.Infof("not mapped approver: %s", approver)
		}
	}

	if len(icons) == 0 {
		// use the default approve icon by default
		icons[iconApproved] = true
	}

	return icons
}

func (c *command) GetHelp() []bot.Help {
	return c.fetcher.getHelp()
}
