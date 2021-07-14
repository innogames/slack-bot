package vcs

import (
	"testing"
	"time"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/stretchr/testify/assert"
)

func TestInitBranchWatcher(t *testing.T) {
	t.Run("Null fetcher", func(t *testing.T) {
		cfg := &config.Config{}

		branches = []string{
			"release/3.12.23",
		}

		assert.Len(t, GetBranches(), 1)

		ctx := util.NewServerContext()
		go InitBranchWatcher(cfg, ctx)
		time.Sleep(time.Millisecond * 10)
		ctx.StopTheWorld()

		// as a nullFetcher is used -> should be empty now
		assert.Len(t, GetBranches(), 0)
	})

	t.Run("Git", func(t *testing.T) {
		cfg := &config.Config{}
		cfg.BranchLookup.Type = "git"
		cfg.BranchLookup.Repository = "test.git"

		fetcher := createBranchFetcher(cfg)

		assert.Equal(t, "test.git", fetcher.(git).repoURL)
	})

	t.Run("Bitbucket with invalid config", func(t *testing.T) {
		cfg := &config.Config{}
		cfg.BranchLookup.Type = "bitbucket"

		fetcher := createBranchFetcher(cfg)

		// we expect a null-fetcher as we don't have valid bitbucket config
		assert.IsType(t, null{}, fetcher)
	})

	t.Run("Bitbucket", func(t *testing.T) {
		cfg := &config.Config{}
		cfg.BranchLookup.Type = "bitbucket"
		cfg.Bitbucket.Host = "https://bitbucket.example.com"
		cfg.Bitbucket.APIKey = "iamsecret"

		fetcher := createBranchFetcher(cfg)

		assert.IsType(t, &bitbucket{}, fetcher)
	})
}

func TestGetMatchingBranches(t *testing.T) {
	branches = []string{
		"master",
		"feature/PROJ-1234-do-something",
		"feature/PROJ-1234-do-something-hotfix",
		"bugfix/PROJ-1235-fixed",
		"release/3.12.23",
	}

	t.Run("Empty", func(t *testing.T) {
		actual, err := GetMatchingBranch("")
		assert.NotNil(t, err)
		assert.Equal(t, "", actual)
	})

	t.Run("Not found", func(t *testing.T) {
		actual, err := GetMatchingBranch("this-might-be-a-branch")
		assert.Equal(t, "this-might-be-a-branch", actual)
		assert.Nil(t, err)
	})

	t.Run("Not unique", func(t *testing.T) {
		actual, err := GetMatchingBranch("PROJ-1234")
		assert.EqualError(t, err, "multiple branches found: feature/PROJ-1234-do-something, feature/PROJ-1234-do-something-hotfix")
		assert.Equal(t, "", actual)
	})

	t.Run("Test unique branches", func(t *testing.T) {
		actual, err := GetMatchingBranch("master")
		assert.Equal(t, "master", actual)
		assert.Nil(t, err)

		actual, err = GetMatchingBranch("PROJ-1235")
		assert.Equal(t, "bugfix/PROJ-1235-fixed", actual)
		assert.Nil(t, err)

		actual, err = GetMatchingBranch("feature/PROJ-1234-do-something")
		assert.Equal(t, "feature/PROJ-1234-do-something", actual)
		assert.Nil(t, err)
	})
}
