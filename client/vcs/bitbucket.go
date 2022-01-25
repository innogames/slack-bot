package vcs

import (
	bitbucketApi "github.com/gfleury/go-bitbucket-v1"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/pkg/errors"
)

// by default bitbucket will only return the most recent 25 branches...
const bitbucketBranchLimit = 1000

type bitbucket struct {
	client *bitbucketApi.DefaultApiService
	cfg    config.Bitbucket
}

// LoadBranches will load the branches from a stash/bitbucket server
func (f *bitbucket) LoadBranches() (branchNames []string, err error) {
	branchesRaw, err := f.client.GetBranches(f.cfg.Project, f.cfg.Repository, map[string]interface{}{
		"limit": bitbucketBranchLimit,
	})
	if err != nil {
		return
	}

	branchesRaw.Body.Close()

	branches, err := bitbucketApi.GetBranchesResponse(branchesRaw)
	if err != nil {
		return branchNames, errors.Wrap(err, "Can't load branched from Bitbucket")
	}

	branchNames = make([]string, 0, len(branches))
	for _, branch := range branches {
		branchNames = append(branchNames, branch.DisplayID)
	}

	return
}
