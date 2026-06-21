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

// a successful read from the persistent layer should be cached in the memory layer
func TestChainStorageReadCachesToMemory(t *testing.T) {
	persistent := newMemoryStorage()
	memory := newMemoryStorage()
	chain := NewChainStorage(persistent, memory)

	require.NoError(t, persistent.Write("collection", "key", "value"))

	// the value only lives in the persistent layer so far
	var fromMemory string
	require.Error(t, memory.Read("collection", "key", &fromMemory))

	// reading through the chain promotes it into the memory layer
	var fromChain string
	require.NoError(t, chain.Read("collection", "key", &fromChain))
	require.Equal(t, "value", fromChain)

	require.NoError(t, memory.Read("collection", "key", &fromMemory))
	require.Equal(t, "value", fromMemory)
}

// a failing read must not poison the memory layer with a zero value
func TestChainStorageFailedReadDoesNotCache(t *testing.T) {
	persistent := newMemoryStorage()
	memory := newMemoryStorage()
	chain := NewChainStorage(persistent, memory)

	var v string
	require.Error(t, chain.Read("collection", "missing", &v))

	require.Error(t, memory.Read("collection", "missing", &v))
}
