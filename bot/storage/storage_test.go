package storage

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStorage(t *testing.T) {
	after := MockStorage()
	defer after()

	var err error
	var value string
	t.Run("read/write", func(t *testing.T) {
		err = Write("collection", "test", "1")
		assert.Nil(t, err)
		err = Write("collection", "test2", "1")
		assert.Nil(t, err)

		// read invalid data
		err = Read("collection1", "test", &value)
		assert.Error(t, err)
		err = Read("collection", "test1", &value)
		assert.Error(t, err)
		assert.Equal(t, "", value)

		// read valid data
		err = Read("collection", "test", &value)
		assert.Nil(t, err)
		assert.Equal(t, "1", value)

		// delete
		err = Delete("collection", "test")
		assert.Nil(t, err)
		err = Read("collection1", "test", &value)
		assert.Error(t, err)

		// delete all
		err := DeleteAll()
		assert.Nil(t, err)
		err = Read("collection", "test2", &value)
		assert.Error(t, err)
	})
}
