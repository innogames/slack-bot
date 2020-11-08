package storage

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func testStorage(t *testing.T, storage storage) {
	var err error
	var stringValue string
	var intValue int
	var mapValue map[int]float32

	const collection = "collection"

	err = storage.Write(collection, "test-string", "1")
	assert.Nil(t, err)
	err = storage.Write(collection, "test-int", 2)
	assert.Nil(t, err)
	err = storage.Write(collection, "test-map", map[int]float32{12: 5.5})
	assert.Nil(t, err)

	// read invalid data
	err = storage.Read("collection1", "test", &stringValue)
	assert.Error(t, err)
	err = storage.Read(collection, "test1", &stringValue)
	assert.Error(t, err)
	assert.Equal(t, "", stringValue)

	// ket keys all
	keys, err := storage.GetKeys(collection)
	assert.Nil(t, err)
	assert.Len(t, keys, 3)
	//assert.Equal(t, []string{"test-string", "test-int", "test-map"}, keys)

	keys, err = storage.GetKeys("invalid-collection")
	assert.Error(t, err)
	assert.Len(t, keys, 0)

	// read valid data
	err = storage.Read(collection, "test-string", &stringValue)
	assert.Nil(t, err)
	assert.Equal(t, "1", stringValue)
	err = storage.Read(collection, "test-int", &intValue)
	assert.Nil(t, err)
	assert.Equal(t, 2, intValue)
	err = storage.Read(collection, "test-map", &mapValue)
	assert.Nil(t, err)
	assert.Equal(t, map[int]float32{12: 5.5}, mapValue)

	// delete
	err = storage.Delete(collection, "test-string")
	assert.Nil(t, err)
	err = storage.Delete(collection, "test-int")
	assert.Nil(t, err)
	err = storage.Delete(collection, "test-map")
	assert.Nil(t, err)
	err = storage.Read("collection1", "test-map", &stringValue)
	assert.Error(t, err)

	keys, err = storage.GetKeys(collection)
	assert.Error(t, err)
	assert.Len(t, keys, 0)

	keys, err = GetKeys("../")
	assert.EqualError(t, err, "invalid storage key: ../")
	assert.Len(t, keys, 0)
}

func TestStorage(t *testing.T) {

	t.Run("validate keys", func(t *testing.T) {
		var err error
		err = validateKey("valid", "also-val-id")
		assert.Nil(t, err)

		err = validateKey("valid", "not#valid")
		assert.EqualError(t, err, "invalid storage key: not#valid")

		err = validateKey("valid", "../../passwd")
		assert.EqualError(t, err, "invalid storage key: ../../passwd")

		err = validateKey("")
		assert.EqualError(t, err, "invalid storage key: ")
	})

	t.Run("test init storage", func(t *testing.T) {
		storage := getStorage()

		assert.IsType(t, &memoryStorage{}, storage)

		err := InitStorage(".")
		storage = getStorage()
		assert.NoError(t, err)
		assert.IsType(t, &fileStorage{}, storage)

		err = InitStorage("")
		storage = getStorage()
		assert.NoError(t, err)
		assert.IsType(t, &memoryStorage{}, storage)

		fileStorage, _ := newFileStorage(".")
		SetStorage(fileStorage)
		assert.NoError(t, err)
		assert.Equal(t, fileStorage, getStorage())
	})
}
