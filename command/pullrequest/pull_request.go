package pullrequest

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/command/queue"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"net"
	"time"
)

const (
	iconInReview        = "eyes"
	iconApproved        = "white_check_mark"
	iconMerged          = "twisted_rightwards_arrows"
	iconDeclined        = "x"
	iconBuildFailed     = "fire"
	iconBuildRunning    = "arrows_counterclockwise"
	iconError           = "x"
	minCheckInterval    = time.Second * 30
	maxCheckInterval    = time.Minute * 3
	maxConnectionErrors = 10
)

type buildStatus int8

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
	fetcher     fetcher
	regexp      string
}

type pullRequest struct {
	name        string
	declined    bool
	merged      bool
	closed      bool
	inReview    bool
	buildStatus buildStatus
	approvers   []string
}

func (c command) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher(c.regexp, c.Execute)
}

func (c command) Execute(match matcher.Result, message msg.Message) {
	_, err := c.fetcher.getPullRequest(match)

	if err != nil {
		c.slackClient.ReplyError(message, err)
		return
	}

	go c.watch(match, message)
}

func (c command) watch(match matcher.Result, message msg.Message) {
	msgRef := slack.NewRefToMessage(message.Channel, message.Timestamp)

	hasApproval := false
	connectionErrors := 0

	done := queue.AddRunningCommand(message, message.Text)
	defer func() {
		done <- true
	}()

	delay := util.GetIncreasingDelay(minCheckInterval, maxCheckInterval)
	currentReactions := c.getOwnReactions(msgRef)

	var pr pullRequest
	var err error

	for {
		pr, err = c.fetcher.getPullRequest(match)
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
				time.Sleep(minCheckInterval)
				continue
			}

			connectionErrors++
			if connectionErrors > maxConnectionErrors {
				// reply error in new thread
				c.slackClient.ReplyError(
					message,
					errors.Wrapf(err, "Error while fetching PR data %d times in a row", connectionErrors),
				)
				c.slackClient.AddReaction(iconError, message)
				return
			}
			continue
		}
		connectionErrors = 0

		// add approved reaction(s)
		if len(pr.approvers) > 0 {
			for icon := range c.getApproveIcons(pr.approvers) {
				c.addReaction(currentReactions, icon, message)
			}

			hasApproval = true
		}

		// add :eyes: when someone is reviewing the PR but nobody approved it yet
		if pr.inReview && !hasApproval && !pr.merged {
			c.addReaction(currentReactions, iconInReview, message)
		} else {
			c.removeReaction(currentReactions, iconInReview, message)
		}

		c.processBuildStatus(pr, currentReactions, message)

		// add merged reaction
		if pr.merged || pr.closed {
			c.addReaction(currentReactions, iconMerged, message)

			return
		}

		// add declined reaction
		if pr.declined {
			c.removeReaction(currentReactions, iconApproved, message)
			c.addReaction(currentReactions, iconDeclined, message)

			return
		}

		time.Sleep(delay.GetNextDelay())
	}
}

// add reactions based on the build status:
// running: iconBuildRunning
// failed: iconBuildFailed
func (c command) processBuildStatus(pr pullRequest, currentReactions map[string]bool, message msg.Message) {
	// monitor build status
	if pr.buildStatus == buildStatusFailed {
		c.addReaction(currentReactions, iconBuildFailed, message)
	} else {
		c.removeReaction(currentReactions, iconBuildFailed, message)
	}

	if pr.buildStatus == buildStatusRunning {
		c.addReaction(currentReactions, iconBuildRunning, message)
	} else {
		c.removeReaction(currentReactions, iconBuildRunning, message)
	}
}

// get the current reactions in the given message which got created by this bot user
func (c command) getOwnReactions(msgRef slack.ItemRef) map[string]bool {
	currentReactions := make(map[string]bool)
	reactions, _ := c.slackClient.GetReactions(msgRef, slack.NewGetReactionsParameters())

	for _, reaction := range reactions {
		for _, user := range reaction.Users {
			if user == client.BotUserID {
				currentReactions[reaction.Name] = true
				break
			}
		}
	}

	return currentReactions
}

func (c command) removeReaction(currentReactions map[string]bool, icon string, message msg.Message) {
	if ok := currentReactions[icon]; !ok {
		// already removed
		return
	}

	delete(currentReactions, icon)
	c.slackClient.RemoveReaction(icon, message)
}

func (c *command) addReaction(currentReactions map[string]bool, icon string, message msg.Message) {
	if _, ok := currentReactions[icon]; ok {
		// already added
		return
	}

	currentReactions[icon] = true

	c.slackClient.AddReaction(icon, message)
}

// generates a map of all icons for the given approvers list. If there is no special mapping, it returns the default icon
func (c command) getApproveIcons(approvers []string) map[string]bool {
	icons := make(map[string]bool)

	for _, approver := range approvers {
		if icon, ok := c.cfg.CustomApproveReaction[approver]; ok {
			icons[icon] = true
		} else {
			log.Infof("not mapped approver: %s", approver)
		}
	}

	if len(icons) == 0 {
		// use the default approve icon by default
		icons[iconApproved] = true
	}

	return icons
}

func (c command) GetHelp() []bot.Help {
	return c.fetcher.getHelp()
}
