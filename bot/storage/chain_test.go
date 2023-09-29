package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChainStorage(t *testing.T) {
	dir := "./test_chain_storage"
	defer os.RemoveAll(dir)

	fileStorage, err := newFileStorage(dir)
	assert.Nil(t, err)

	storage := NewChainStorage(fileStorage, newMemoryStorage())

	t.Run("test chain storage", func(t *testing.T) {
		testStorage(t, storage)
	})
}
