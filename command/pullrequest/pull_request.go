package pullrequest

import (
	"net"
	"text/template"
	"time"

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
)

const (
	iconInReview     = "eyes"
	iconApproved     = "white_check_mark"
	iconMerged       = "twisted_rightwards_arrows"
	iconClosed       = "x"
	iconBuildFailed  = "fire"
	iconBuildRunning = "arrows_counterclockwise"
	iconError        = "x"
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

	// list of usernames which approved the PR
	Approvers []string
}

func (c command) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher(c.regexp, c.Execute)
}

func (c command) Execute(match matcher.Result, message msg.Message) {
	_, err := c.fetcher.getPullRequest(match)

	if err != nil {
		c.ReplyError(message, err)
		return
	}

	go c.watch(match, message)
}

func (c command) watch(match matcher.Result, message msg.Message) {
	msgRef := slack.NewRefToMessage(message.Channel, message.Timestamp)
	currentErrorCount := 0
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
				c.AddReaction(iconError, message)
				return
			}

			// wait some time before the next retry...might be some server restart or whatever
			time.Sleep(maxCheckInterval)

			continue
		}
		currentErrorCount = 0

		c.setPRReactions(pr, currentReactions, message)

		// stop watching!
		if pr.Status == prStatusClosed || pr.Status == prStatusMerged {
			return
		}

		time.Sleep(delay.GetNextDelay())
	}
}

func (c command) setPRReactions(pr pullRequest, currentReactions map[string]bool, message msg.Ref) {
	hasApproval := false

	// add approved reaction(s)
	if len(pr.Approvers) > 0 {
		for icon := range c.getApproveIcons(pr.Approvers) {
			c.addReaction(currentReactions, icon, message)
		}

		hasApproval = true
	} else {
		c.removeReaction(currentReactions, iconApproved, message)
	}

	c.processBuildStatus(pr, currentReactions, message)

	// add :eyes: when someone is reviewing the PR but nobody approved it yet
	if pr.Status == prStatusInReview && !hasApproval {
		c.addReaction(currentReactions, iconInReview, message)
	} else {
		c.removeReaction(currentReactions, iconInReview, message)
	}

	if pr.Status == prStatusMerged {
		c.addReaction(currentReactions, iconMerged, message)
	} else if pr.Status == prStatusClosed {
		c.removeReaction(currentReactions, iconApproved, message)
		c.addReaction(currentReactions, iconClosed, message)
	}
}

// add reactions based on the build Status:
// running: iconBuildRunning
// failed: iconBuildFailed
func (c command) processBuildStatus(pr pullRequest, currentReactions map[string]bool, message msg.Ref) {
	// monitor build Status
	if pr.BuildStatus == buildStatusFailed {
		c.addReaction(currentReactions, iconBuildFailed, message)
	} else {
		c.removeReaction(currentReactions, iconBuildFailed, message)
	}

	if pr.BuildStatus == buildStatusRunning {
		c.addReaction(currentReactions, iconBuildRunning, message)
	} else {
		c.removeReaction(currentReactions, iconBuildRunning, message)
	}
}

// get the current reactions in the given message which got created by this bot user
func (c command) getOwnReactions(msgRef slack.ItemRef) map[string]bool {
	currentReactions := make(map[string]bool)
	reactions, _ := c.GetReactions(msgRef, slack.NewGetReactionsParameters())

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

func (c command) removeReaction(currentReactions map[string]bool, icon string, message msg.Ref) {
	if ok := currentReactions[icon]; !ok {
		// already removed
		return
	}

	delete(currentReactions, icon)
	c.RemoveReaction(icon, message)
}

func (c *command) addReaction(currentReactions map[string]bool, icon string, message msg.Ref) {
	if _, ok := currentReactions[icon]; ok {
		// already added
		return
	}

	currentReactions[icon] = true

	c.AddReaction(icon, message)
}

// generates a map of all icons for the given Approvers list. If there is no special mapping, it returns the default icon
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

func (c command) GetTemplateFunction() template.FuncMap {
	if functions, ok := c.fetcher.(util.TemplateFunctionProvider); ok {
		return functions.GetTemplateFunction()
	}

	return template.FuncMap{}
}

func (c command) GetHelp() []bot.Help {
	return c.fetcher.getHelp()
}
