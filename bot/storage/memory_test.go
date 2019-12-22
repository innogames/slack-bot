package storage

import (
	"testing"
)

func TestMemoryStorage(t *testing.T) {
	storage := newMemoryStorage()

	t.Run("test memory", func(t *testing.T) {
		testStorage(t, storage)
	})
}
