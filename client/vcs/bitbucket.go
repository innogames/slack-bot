package vcs

import (
	"github.com/innogames/slack-bot/bot/config"
	"github.com/xoom/stash"
)

// todo switch to https://github.com/ktrysmt/go-bitbucket
type bitbucket struct {
	client stash.Stash
	cfg    config.Bitbucket
}

// LoadBranches will load the branches from a stash/bitbucket server
func (f bitbucket) LoadBranches() ([]string, error) {
	var branchNames []string

	branches, err := f.client.GetBranches(f.cfg.Project, f.cfg.Repository)
	if err != nil {
		return branchNames, err
	}

	branchNames = make([]string, 0, len(branches))
	for branchName := range branches {
		branchNames = append(branchNames, branchName)
	}

	return branchNames, nil
}
