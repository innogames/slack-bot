package vcs

import (
	bitbucketApi "github.com/gfleury/go-bitbucket-v1"
	"github.com/innogames/slack-bot/bot/config"
)

type bitbucket struct {
	client *bitbucketApi.APIClient
	cfg    config.Bitbucket
}

// LoadBranches will load the branches from a stash/bitbucket server
func (f bitbucket) LoadBranches() ([]string, error) {
	var branchNames []string

	branchesRaw, err := f.client.DefaultApi.GetBranches(f.cfg.Project, f.cfg.Repository, nil)
	if err != nil {
		return branchNames, err
	}

	branches, err := bitbucketApi.GetBranchesResponse(branchesRaw)
	if err != nil {
		return branchNames, err
	}

	branchNames = make([]string, 0, len(branches))
	for _, branch := range branches {
		branchNames = append(branchNames, branch.DisplayID)
	}

	return branchNames, nil
}
