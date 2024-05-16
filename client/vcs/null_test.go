package vcs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNullLoader(t *testing.T) {
	fetcher := &null{}
	t.Run("Load branches with null loader", func(t *testing.T) {
		branches, err := fetcher.LoadBranches()
		assert.Len(t, branches, 0)
		require.NoError(t, err)
	})
}
