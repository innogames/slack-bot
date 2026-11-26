package vcs

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitBranchWatcher(t *testing.T) {
	t.Run("Null fetcher", func(t *testing.T) {
		cfg := &config.Config{}
		cfg.BranchLookup.UpdateInterval = time.Second

		branches = []string{
			"release/3.12.23",
		}

		assert.Len(t, GetBranches(), 1)

		ctx := util.NewServerContext()
		go InitBranchWatcher(cfg, ctx)
		time.Sleep(time.Millisecond * 10)
		ctx.StopTheWorld()

		// as a nullFetcher is used -> should be empty now
		assert.Empty(t, GetBranches())
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

type testFetcher struct {
	branches []string
	err      error
}

func (f testFetcher) LoadBranches() ([]string, error) {
	return f.branches, f.err
}

func TestRefreshBranches(t *testing.T) {
	oldBranches := branches
	defer func() { branches = oldBranches }()

	t.Run("Successful refresh replaces the branches", func(t *testing.T) {
		branches = []string{"master"}

		refreshBranches(testFetcher{branches: []string{"master", "release/3.12.23"}})

		assert.Equal(t, []string{"master", "release/3.12.23"}, GetBranches())
	})

	t.Run("Failed refresh keeps the known branches", func(t *testing.T) {
		branches = []string{"master", "feature/PROJ-1234-do-something"}

		refreshBranches(testFetcher{err: errors.New("connection refused")})

		assert.Equal(t, []string{"master", "feature/PROJ-1234-do-something"}, GetBranches())

		// the lookup still resolves partial names instead of passing them through
		actual, err := GetMatchingBranch("PROJ-1234")
		require.NoError(t, err)
		assert.Equal(t, "feature/PROJ-1234-do-something", actual)
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
		// empty branch is passed through to jenkins instead of fuzzy-matching every branch
		actual, err := GetMatchingBranch("")
		require.NoError(t, err)
		assert.Empty(t, actual)

		actual, err = GetMatchingBranch("   ")
		require.NoError(t, err)
		assert.Empty(t, actual)
	})

	t.Run("Not found", func(t *testing.T) {
		actual, err := GetMatchingBranch("this-might-be-a-branch")
		assert.Equal(t, "this-might-be-a-branch", actual)
		require.NoError(t, err)
	})

	t.Run("Not unique", func(t *testing.T) {
		actual, err := GetMatchingBranch("PROJ-1234")
		require.EqualError(t, err, "multiple branches found: feature/PROJ-1234-do-something, feature/PROJ-1234-do-something-hotfix")
		assert.Empty(t, actual)
	})

	t.Run("Not unique with truncated list", func(t *testing.T) {
		manyBranches := make([]string, 0, 15)
		for i := range 15 {
			manyBranches = append(manyBranches, fmt.Sprintf("feature/PROJ-99-variant-%02d", i))
		}
		oldBranches := branches
		branches = manyBranches
		defer func() { branches = oldBranches }()

		actual, err := GetMatchingBranch("PROJ-99")
		require.Error(t, err)
		assert.Empty(t, actual)
		assert.Contains(t, err.Error(), "(and 5 more)")
		assert.NotContains(t, err.Error(), "variant-10")
	})

	t.Run("Test unique branches", func(t *testing.T) {
		actual, err := GetMatchingBranch("master")
		assert.Equal(t, "master", actual)
		require.NoError(t, err)

		actual, err = GetMatchingBranch("PROJ-1235")
		assert.Equal(t, "bugfix/PROJ-1235-fixed", actual)
		require.NoError(t, err)

		actual, err = GetMatchingBranch("feature/PROJ-1234-do-something")
		assert.Equal(t, "feature/PROJ-1234-do-something", actual)
		require.NoError(t, err)
	})
}
