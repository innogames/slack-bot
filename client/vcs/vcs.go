package vcs

import (
	"fmt"
	"strings"
	"time"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/innogames/slack-bot/v2/client"
	log "github.com/sirupsen/logrus"
)

const branchFetchInterval = time.Minute * 2

// cached list of branch names
var branches []string

// BranchFetcher loads a list of all available branch names in a repository
type BranchFetcher interface {
	LoadBranches() ([]string, error)
}

// InitBranchWatcher will load the current branches each X from the configured VCS -> e.g. used for branch lookup for Jenkins parameters
func InitBranchWatcher(cfg *config.Config, ctx *util.ServerContext) {
	ctx.RegisterChild()
	defer ctx.ChildDone()

	var err error
	fetcher := createBranchFetcher(cfg)

	// load branch list using startup
	branches, err = fetcher.LoadBranches()
	if err != nil {
		log.Error(err)
	}

	ticker := time.NewTicker(branchFetchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			branches, err = fetcher.LoadBranches()
			if err != nil {
				log.Error(err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// GetBranches returns a list of currently known branches
func GetBranches() []string {
	return branches
}

// GetMatchingBranch does a fuzzy search on all loaded branches. If there are multiple matching branches, it fails.
func GetMatchingBranch(input string) (string, error) {
	var foundBranches []string

	loweredInput := strings.ToLower(input)
	for _, branch := range GetBranches() {
		loweredBranch := strings.ToLower(branch)
		if loweredBranch == loweredInput {
			return input, nil
		} else if strings.Contains(loweredBranch, loweredInput) {
			foundBranches = append(foundBranches, branch)
		}
	}

	if len(foundBranches) > 1 {
		return "", fmt.Errorf("multiple branches found: %s", strings.Join(foundBranches, ", "))
	} else if len(foundBranches) == 1 {
		return foundBranches[0], nil
	}

	log.Errorf("Branch not found: %s. We have %d known branches", input, len(branches))

	// branch not found in local list, but maybe it was created recently -> let's try it if jenkins accept it
	return input, nil
}

func createBranchFetcher(cfg *config.Config) BranchFetcher {
	switch cfg.BranchLookup.Type {
	case "stash", "bitbucket":
		bitbucketClient, err := client.GetBitbucketClient(cfg.Bitbucket)
		if err != nil {
			log.Errorf("Cannot init Bitbucket client: %s", err)
			return null{}
		}

		return &bitbucket{bitbucketClient, cfg.Bitbucket}
	case "git":
		return git{cfg.BranchLookup.Repository}
	default:
		return null{}
	}
}
