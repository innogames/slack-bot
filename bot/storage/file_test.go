package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileStorage(t *testing.T) {
	dir := "./test_storage"
	defer os.RemoveAll(dir)

	storage, err := newFileStorage(dir)
	assert.Nil(t, err)

	t.Run("test file", func(t *testing.T) {
		testStorage(t, storage)
	})
}
