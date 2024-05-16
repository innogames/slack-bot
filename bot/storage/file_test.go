package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileStorage(t *testing.T) {
	dir := "./test_storage"
	defer os.RemoveAll(dir)

	storage, err := newFileStorage(dir)
	require.NoError(t, err)

	t.Run("test file", func(t *testing.T) {
		testStorage(t, storage)
	})
}
