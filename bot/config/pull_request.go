package config

// PullRequest special configuration to change the pull request behavior
type PullRequest struct {
	// overwrite reactions, default ones, see default.go
	Reactions PullRequestReactions `mapstructure:"reactions"`

	// able to set a custom "approved" reactions to see directly who or which component/department approved a pullrequest
	CustomApproveReaction map[string]string `mapstructure:"custom_approve_reaction"`
}

type PullRequestReactions struct {
	InReview     string `mapstructure:"in_review"`
	Approved     string `mapstructure:"approved"`
	Merged       string `mapstructure:"merged"`
	Closed       string `mapstructure:"closed"`
	BuildFailed  string `mapstructure:"build_failed"`
	BuildRunning string `mapstructure:"build_running"`
	Error        string `mapstructure:"error"`
}
