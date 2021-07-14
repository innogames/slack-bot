package vcs

import (
	bitbucketApi "github.com/gfleury/go-bitbucket-v1"
	"github.com/innogames/slack-bot.v2/bot/config"
)

type bitbucket struct {
	client *bitbucketApi.DefaultApiService
	cfg    config.Bitbucket
}

// LoadBranches will load the branches from a stash/bitbucket server
func (f *bitbucket) LoadBranches() (branchNames []string, err error) {
	branchesRaw, err := f.client.GetBranches(f.cfg.Project, f.cfg.Repository, nil)
	if err != nil {
		return
	}

	branchesRaw.Body.Close()

	branches, err := bitbucketApi.GetBranchesResponse(branchesRaw)
	if err != nil {
		return branchNames, err
	}

	branchNames = make([]string, 0, len(branches))
	for _, branch := range branches {
		branchNames = append(branchNames, branch.DisplayID)
	}

	return
}
