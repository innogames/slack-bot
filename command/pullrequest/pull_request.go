package pullrequest

import (
	"fmt"
	"net"
	"text/template"
	"time"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/innogames/slack-bot/v2/command/queue"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

const (
	minCheckInterval = time.Second * 30
	maxCheckInterval = time.Minute * 3
	maxErrors        = 5 // number of max errors in a row before aborting the PR watcher
)

type buildStatus int8

const (
	buildStatusUnknown buildStatus = iota
	buildStatusSuccess
	buildStatusFailed
	buildStatusRunning
)

type prStatus uint8

const (
	prStatusOpen prStatus = iota
	prStatusInReview
	prStatusMerged
	prStatusClosed
)

type reactionMap map[util.Reaction]bool

type fetcher interface {
	getPullRequest(match matcher.Result) (pullRequest, error)
	getHelp() []bot.Help
}

type command struct {
	bot.BaseCommand
	cfg     config.PullRequest
	fetcher fetcher
	regexp  string
}

type pullRequest struct {
	// title/name of the PR
	Name   string
	Status prStatus

	// status of a related CI build
	BuildStatus buildStatus

	// author of the PR
	Author string

	// link to the PR
	Link string

	// list of usernames which approved the PR
	Approvers []string
}

func (c command) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher(c.regexp, c.execute)
}

func (c command) execute(match matcher.Result, message msg.Message) {
	_, err := c.fetcher.getPullRequest(match)
	if err != nil {
		c.AddReaction(c.cfg.Reactions.Error, message)
		c.ReplyError(message, err)
		return
	}

	go c.watch(match, message)
}

type pullRequestWatch struct {
	DidNotifyMergeable bool
	PullRequest        pullRequest
	LastBuildStatus    buildStatus
	PullRequestStatus  prStatus
}

func (c command) watch(match matcher.Result, message msg.Message) {
	msgRef := slack.NewRefToMessage(message.Channel, message.Timestamp)
	currentErrorCount := 0

	runningCommand := queue.AddRunningCommand(message, message.Text)
	defer runningCommand.Done()

	delay := util.GetIncreasingDelay(minCheckInterval, maxCheckInterval)
	currentReactions := c.getOwnReactions(msgRef)

	var prw pullRequestWatch
	var err error

	for {
		prw.PullRequest, err = c.fetcher.getPullRequest(match)

		// something failed while loading the PR data...retry if it was temporary, else quit watching
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
				time.Sleep(maxCheckInterval)
				continue
			}

			currentErrorCount++
			if currentErrorCount > maxErrors {
				// reply error in new thread
				c.ReplyError(
					message,
					errors.Wrapf(err, "Error while fetching PR data %d times in a row", currentErrorCount),
				)
				c.AddReaction(c.cfg.Reactions.Error, message)
				return
			}

			// wait some time before the next retry...might be some server restart or whatever
			time.Sleep(maxCheckInterval)

			continue
		}
		currentErrorCount = 0

		c.setPRReactions(prw.PullRequest, currentReactions, message)

		c.notifyBuildStatus(&prw)

		c.notifyPullRequestStatus(&prw)

		// stop watching!
		if prw.PullRequest.Status == prStatusClosed || prw.PullRequest.Status == prStatusMerged {
			return
		}

		time.Sleep(delay.GetNextDelay())
	}
}

func (c command) setPRReactions(pr pullRequest, currentReactions reactionMap, message msg.Ref) {
	hasApproval := false

	// add approved reaction(s)
	if len(pr.Approvers) > 0 {
		for icon := range c.getApproveReactions(pr.Approvers) {
			c.addReaction(currentReactions, icon, message)
		}

		hasApproval = true
	} else {
		c.removeReaction(currentReactions, c.cfg.Reactions.Approved, message)
	}

	c.processBuildStatus(pr, currentReactions, message)

	// add :eyes: when someone is reviewing the PR but nobody approved it yet
	if pr.Status == prStatusInReview && !hasApproval {
		c.addReaction(currentReactions, c.cfg.Reactions.InReview, message)
	} else {
		c.removeReaction(currentReactions, c.cfg.Reactions.InReview, message)
	}

	if pr.Status == prStatusMerged {
		c.addReaction(currentReactions, c.cfg.Reactions.Merged, message)
	} else if pr.Status == prStatusClosed {
		c.removeReaction(currentReactions, c.cfg.Reactions.Approved, message)
		c.addReaction(currentReactions, c.cfg.Reactions.Closed, message)
	}
}

// add reactions based on the build Status:
// running: iconBuildRunning
// failed: iconBuildFailed
func (c command) processBuildStatus(pr pullRequest, currentReactions reactionMap, message msg.Ref) {
	// monitor build Status
	if pr.BuildStatus == buildStatusFailed {
		c.addReaction(currentReactions, c.cfg.Reactions.BuildFailed, message)
	} else {
		c.removeReaction(currentReactions, c.cfg.Reactions.BuildFailed, message)
	}

	if pr.BuildStatus == buildStatusRunning {
		c.addReaction(currentReactions, c.cfg.Reactions.BuildRunning, message)
	} else {
		c.removeReaction(currentReactions, c.cfg.Reactions.BuildRunning, message)
	}
}

// get the current reactions in the given message which got created by this bot user
func (c command) getOwnReactions(msgRef slack.ItemRef) reactionMap {
	currentReactions := make(reactionMap)
	reactions, _ := c.GetReactions(msgRef, slack.NewGetReactionsParameters())

	for _, reaction := range reactions {
		for _, user := range reaction.Users {
			if user == client.AuthResponse.UserID {
				currentReactions[util.Reaction(reaction.Name)] = true
				break
			}
		}
	}

	return currentReactions
}

func (c command) removeReaction(currentReactions reactionMap, reaction util.Reaction, message msg.Ref) {
	if ok := currentReactions[reaction]; !ok {
		// already removed
		return
	}

	delete(currentReactions, reaction)
	c.RemoveReaction(reaction, message)
}

func (c *command) addReaction(currentReactions reactionMap, reaction util.Reaction, message msg.Ref) {
	if _, ok := currentReactions[reaction]; ok {
		// already added
		return
	}

	currentReactions[reaction] = true

	c.AddReaction(reaction, message)
}

// generates a map of all reactions for the given Approvers list. If there is no special mapping, it returns the default icon
func (c command) getApproveReactions(approvers []string) reactionMap {
	reactions := make(reactionMap)

	for _, approver := range approvers {
		if reaction, ok := c.cfg.CustomApproveReaction[approver]; ok {
			reactions[reaction] = true
		} else {
			log.Infof("not mapped approver: %s", approver)
		}
	}

	if len(reactions) == 0 {
		// use the default approve icon by default
		reactions[c.cfg.Reactions.Approved] = true
	}

	return reactions
}

func getPRLinkMessage(prw *pullRequestWatch) string {
	if len(prw.PullRequest.Link) > 0 {
		return fmt.Sprintf("\nYou can check it here: %s", prw.PullRequest.Link)
	}
	return ""
}

func (c command) getBuildStatusIcon(pr pullRequest) string {
	switch pr.BuildStatus {
	case buildStatusRunning:
		return fmt.Sprintf(":%s:", c.cfg.Reactions.BuildRunning)
	case buildStatusSuccess:
		return fmt.Sprintf(":%s:", c.cfg.Reactions.BuildSuccess)
	case buildStatusFailed:
		return fmt.Sprintf(":%s:", c.cfg.Reactions.BuildFailed)
	case buildStatusUnknown:
	}
	return ""
}

func (c command) notifyBuildStatus(prw *pullRequestWatch) {
	if len(prw.PullRequest.Author) == 0 {
		// no author to notify
		return
	}

	if prw.PullRequest.BuildStatus == prw.LastBuildStatus {
		// build status didn't change
		return
	}

	prw.LastBuildStatus = prw.PullRequest.BuildStatus

	if c.cfg.Notifications.BuildStatusInProgress && prw.LastBuildStatus == buildStatusRunning {
		c.sendPrivateMessage(prw.PullRequest.Author, "PR '%s'\nBuild Status: Started %s%s", prw.PullRequest.Name, c.getBuildStatusIcon(prw.PullRequest), getPRLinkMessage(prw))
	}
	if c.cfg.Notifications.BuildStatusSuccess && prw.PullRequest.BuildStatus == buildStatusSuccess {
		c.sendPrivateMessage(prw.PullRequest.Author, "PR '%s'\nBuild Status: Success %s%s", prw.PullRequest.Name, c.getBuildStatusIcon(prw.PullRequest), getPRLinkMessage(prw))
	}
	if c.cfg.Notifications.BuildStatusFailed && prw.PullRequest.BuildStatus == buildStatusFailed {
		c.sendPrivateMessage(prw.PullRequest.Author, "PR '%s'\nBuild Status: Failed %s%s", prw.PullRequest.Name, c.getBuildStatusIcon(prw.PullRequest), getPRLinkMessage(prw))
	}
}

func (c command) notifyPullRequestStatus(prw *pullRequestWatch) {
	if prw.DidNotifyMergeable {
		return
	}

	if len(prw.PullRequest.Author) == 0 {
		// no author to notify
		return
	}

	if len(prw.PullRequest.Approvers) == 0 {
		// prw isn't approved
		return
	}

	if prw.PullRequest.BuildStatus != buildStatusSuccess {
		// has still running or failed builds
		return
	}

	prw.DidNotifyMergeable = true

	c.sendPrivateMessage(prw.PullRequest.Author, "Your PR '%s' is ready to merge!%s", prw.PullRequest.Name, getPRLinkMessage(prw))
}

func (c command) sendPrivateMessage(username string, format string, parameter ...any) {
	message := fmt.Sprintf(format, parameter...)
	c.SendToUser(username, message)
}

func (c command) GetTemplateFunction() template.FuncMap {
	if functions, ok := c.fetcher.(util.TemplateFunctionProvider); ok {
		return functions.GetTemplateFunction()
	}

	return template.FuncMap{}
}

func (c command) GetHelp() []bot.Help {
	return c.fetcher.getHelp()
}
