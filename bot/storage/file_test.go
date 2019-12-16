package storage

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestFileStorage(t *testing.T) {
	dir := "./test_storage"
	defer os.RemoveAll(dir)

	storage, err := newFileStorage(dir)
	assert.Nil(t, err)

	t.Run("test all", func(t *testing.T) {
		testStorage(t, storage)
	})
}
