package vcs

import (
	"fmt"
	"github.com/innogames/slack-bot/client"
	"strings"
	"time"

	"github.com/innogames/slack-bot/bot/config"
	"github.com/sirupsen/logrus"
)

const branchFetchInterval = time.Minute * 2

// cached list of branch names
var branches []string
var logger *logrus.Logger

// BranchFetcher loads a list of all available branch names in a repository
type BranchFetcher interface {
	LoadBranches() ([]string, error)
}

// InitBranchWatcher will load the current branches each X from the configured VCS -> e.g. used for branch lookup for Jenkins parameters
func InitBranchWatcher(config config.Config, log *logrus.Logger) {
	logger = log
	go func() {
		fetcher := createBranchFetcher(config)
		for {
			var err error
			branches, err = fetcher.LoadBranches()
			if err != nil {
				logger.Error(err)
			}
			time.Sleep(branchFetchInterval)
		}
	}()
}

// GetMatchingBranch does a fuzzy search on all loaded branches. If there are multiple matching branches, it fails.
func GetMatchingBranch(input string) (string, error) {
	var foundBranches []string

	loweredInput := strings.ToLower(input)
	for _, branch := range branches {
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

	logger.Errorf("Branch not found: %s. We have %d known branches", input, len(branches))

	// branch not found in local list, but maybe it was created recently -> let's try it if jenkins accept it
	return input, nil
}

func createBranchFetcher(cfg config.Config) BranchFetcher {
	switch cfg.BranchLookup.Type {
	case "stash", "bitbucket":
		bitbucketClient, err := client.GetBitbucketClient(cfg.Bitbucket)
		if err != nil {
			logger.Errorf("Cannot init Bitbucket client: %s", err)
		}

		return bitbucket{bitbucketClient, cfg.Bitbucket}
	case "git":
		return git{cfg.BranchLookup.Repository}
	default:
		return null{}
	}
}
