package vcs

import (
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNullLoader(t *testing.T) {
	fetcher := &null{}
	t.Run("Load branches with null loader", func(t *testing.T) {
		branches, err := fetcher.LoadBranches()
		assert.Len(t, branches, 0)
		assert.Nil(t, err)
	})
}

func TestGetMatchingBranches(t *testing.T) {
	logger, _ = test.NewNullLogger()

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
		assert.Equal(t, "multiple branches found: feature/PROJ-1234-do-something, feature/PROJ-1234-do-something-hotfix", err.Error())
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
