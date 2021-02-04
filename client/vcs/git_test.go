package vcs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitLoader(t *testing.T) {
	t.Run("Load branches with slack-bot repoURL", func(t *testing.T) {
		fetcher := &git{
			repoURL: "https://github.com/innogames/slack-bot.git",
		}

		branches, err := fetcher.LoadBranches()
		assert.True(t, len(branches) >= 1)
		assert.Nil(t, err)
	})

	t.Run("Load branches with invalid repoURL", func(t *testing.T) {
		fetcher := &git{
			repoURL: "sdsdsdsd",
		}

		branches, err := fetcher.LoadBranches()

		assert.Empty(t, branches)
		assert.NotNil(t, err)
	})
}
