package config

import "github.com/innogames/slack-bot/v2/bot/util"

// PullRequest special configuration to change the pull request behavior
type PullRequest struct {
	// overwrite reactions, default ones, see default.go
	Reactions PullRequestReactions `mapstructure:"reactions"`

	// able to set a custom "approved" reactions to see directly who or which component/department approved a pullrequest
	CustomApproveReaction map[string]util.Reaction `mapstructure:"custom_approve_reaction"`
}

// PullRequestReactions can be defined in the config.yaml to have custom reactions for pull requests.
// the defaults are defined in default.go
type PullRequestReactions struct {
	InReview     util.Reaction `mapstructure:"in_review"`
	Approved     util.Reaction `mapstructure:"approved"`
	Merged       util.Reaction `mapstructure:"merged"`
	Closed       util.Reaction `mapstructure:"closed"`
	BuildFailed  util.Reaction `mapstructure:"build_failed"`
	BuildRunning util.Reaction `mapstructure:"build_running"`
	Error        util.Reaction `mapstructure:"error"`
}
