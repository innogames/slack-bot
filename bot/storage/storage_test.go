package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testStorage(t *testing.T, storage Storage) {
	t.Helper()

	var err error
	var stringValue string
	var intValue int
	var mapValue map[int]float32

	const collection = "collection"

	err = storage.Write(collection, "test-string", "1")
	require.NoError(t, err)
	err = storage.Write(collection, "test-int", 2)
	require.NoError(t, err)
	err = storage.Write(collection, "test-map", map[int]float32{12: 5.5})
	require.NoError(t, err)

	// read invalid data
	err = storage.Read("collection1", "test", &stringValue)
	require.Error(t, err)
	err = storage.Read(collection, "test1", &stringValue)
	require.Error(t, err)
	assert.Equal(t, "", stringValue)

	// get keys all
	keys, err := storage.GetKeys(collection)
	require.NoError(t, err)
	assert.Len(t, keys, 3)

	keys, err = storage.GetKeys("invalid-collection")
	require.NoError(t, err)
	assert.Len(t, keys, 0)

	// read valid data
	err = storage.Read(collection, "test-string", &stringValue)
	require.NoError(t, err)
	assert.Equal(t, "1", stringValue)
	err = storage.Read(collection, "test-int", &intValue)
	require.NoError(t, err)
	assert.Equal(t, 2, intValue)
	err = storage.Read(collection, "test-map", &mapValue)
	require.NoError(t, err)
	assert.Equal(t, map[int]float32{12: 5.5}, mapValue)

	// expected unmarshall error when accessing wrong data
	err = storage.Read(collection, "test-int", &mapValue)
	assert.EqualError(t, err, "json: cannot unmarshal number into Go value of type map[int]float32")

	// Atomic
	Atomic(func() {
		err = storage.Write(collection, "test-int", 1)
		require.NoError(t, err)
		err = storage.Write(collection, "test-int", 2)
		require.NoError(t, err)
	})
	var expectedInt int
	err = storage.Read(collection, "test-int", &expectedInt)
	assert.Equal(t, 2, expectedInt)

	// delete
	err = storage.Delete(collection, "test-string")
	require.NoError(t, err)
	err = storage.Delete(collection, "test-int")
	require.NoError(t, err)
	err = storage.Delete(collection, "test-map")
	require.NoError(t, err)
	err = storage.Read("collection1", "test-map", &stringValue)
	require.Error(t, err)

	// Atomic
	Atomic(func() {
		err = storage.Write(collection, "test-int2", 1)
		require.NoError(t, err)
		err = storage.Write(collection, "test-int2", 2)
		require.NoError(t, err)
	})
	err = storage.Read(collection, "test-int2", &expectedInt)
	assert.Equal(t, 2, expectedInt)
	require.NoError(t, err)
	err = storage.Delete(collection, "test-int2")
	require.NoError(t, err)

	keys, err = storage.GetKeys(collection)
	require.NoError(t, err)
	assert.Len(t, keys, 0)

	keys, err = GetKeys("../")
	assert.EqualError(t, err, "invalid Storage key: ../")
	assert.Len(t, keys, 0)
}

func TestStorage(t *testing.T) {
	t.Run("validate keys", func(t *testing.T) {
		var err error
		err = validateKey("valid", "also-val-id")
		require.NoError(t, err)

		err = validateKey("valid", "not#valid")
		assert.EqualError(t, err, "invalid Storage key: not#valid")

		err = validateKey("valid", "../../passwd")
		assert.EqualError(t, err, "invalid Storage key: ../../passwd")

		err = validateKey("")
		assert.EqualError(t, err, "invalid Storage key: ")
	})

	t.Run("test init Storage", func(t *testing.T) {
		storage := getStorage()

		assert.IsType(t, &memoryStorage{}, storage)

		err := InitStorage(".")
		storage = getStorage()
		require.NoError(t, err)
		assert.IsType(t, &chainStorage{}, storage)

		err = InitStorage("")
		storage = getStorage()
		require.NoError(t, err)
		assert.IsType(t, &memoryStorage{}, storage)

		fileStorage, _ := newFileStorage(".")
		SetStorage(fileStorage)
		require.NoError(t, err)
		assert.Equal(t, fileStorage, getStorage())
	})

	t.Run("Validate keys", func(t *testing.T) {
		err := Write("../test", "foo", "")
		assert.EqualError(t, err, "invalid Storage key: ../test")

		err = Delete("../test", "foo")
		assert.EqualError(t, err, "invalid Storage key: ../test")

		err = DeleteCollection("../test")
		assert.EqualError(t, err, "invalid Storage key: ../test")

		var i int
		err = Read("../test", "dd", &i)
		assert.EqualError(t, err, "invalid Storage key: ../test")
	})
}
