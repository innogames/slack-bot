package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChainStorage(t *testing.T) {
	dir := "./test_chain_storage"
	defer os.RemoveAll(dir)

	fileStorage, err := newFileStorage(dir)
	require.NoError(t, err)

	storage := NewChainStorage(fileStorage, newMemoryStorage())

	t.Run("test chain storage", func(t *testing.T) {
		testStorage(t, storage)
	})
}
