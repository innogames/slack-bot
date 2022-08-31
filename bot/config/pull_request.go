package config

import "github.com/innogames/slack-bot/v2/bot/util"

// PullRequest special configuration to change the pull request behavior
type PullRequest struct {
	// overwrite reactions, default ones, see default.go
	Reactions PullRequestReactions `mapstructure:"reactions"`

	// enable private notifications for the build status
	Notifications Notifications `mapstructure:"notifications"`

	// able to set a custom "approved" reactions to see directly who or which component/department approved a pullrequest
	CustomApproveReaction map[string]util.Reaction `mapstructure:"custom_approve_reaction"`
}

// Notifications can be defined in the config.yaml to enable notifications for pull request builds.
// the defaults are defined in default.go
type Notifications struct {
	BuildStatusInProgress      bool `mapstructure:"build_status_in_progress"`
	BuildStatusSuccess         bool `mapstructure:"build_status_success"`
	BuildStatusFailed          bool `mapstructure:"build_status_failed"`
	PullRequestStatusMergeable bool `mapstructure:"pr_status_mergeable"`
}

// PullRequestReactions can be defined in the config.yaml to have custom reactions for pull requests.
// the defaults are defined in default.go
type PullRequestReactions struct {
	InReview     util.Reaction `mapstructure:"in_review"`
	Approved     util.Reaction `mapstructure:"approved"`
	Merged       util.Reaction `mapstructure:"merged"`
	Closed       util.Reaction `mapstructure:"closed"`
	BuildSuccess util.Reaction `mapstructure:"build_success"`
	BuildRunning util.Reaction `mapstructure:"build_running"`
	BuildFailed  util.Reaction `mapstructure:"build_failed"`
	Error        util.Reaction `mapstructure:"error"`
}
