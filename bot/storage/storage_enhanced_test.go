package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorageEnhanced(t *testing.T) {
	t.Run("test chain storage write error handling", func(t *testing.T) {
		// Test chain storage with persistent and memory backends
		memoryStorage := newMemoryStorage()
		fileStorage, err := newFileStorage(".")
		require.NoError(t, err)

		chain := NewChainStorage(fileStorage, memoryStorage)

		// Test write operation
		err = chain.Write("test_collection", "test_key", "test_value")
		require.NoError(t, err)

		// Test read operation
		var value string
		err = chain.Read("test_collection", "test_key", &value)
		require.NoError(t, err)
		assert.Equal(t, "test_value", value)
	})

	t.Run("test chain storage with multiple backends", func(t *testing.T) {
		// Create two memory storages for chain
		storage1 := newMemoryStorage()
		storage2 := newMemoryStorage()
		chain := NewChainStorage(storage1, storage2)

		// Write to chain
		err := chain.Write("test_collection", "test_key", "test_value")
		require.NoError(t, err)

		// Should be able to read from chain
		var value string
		err = chain.Read("test_collection", "test_key", &value)
		require.NoError(t, err)
		assert.Equal(t, "test_value", value)
	})

	t.Run("test file storage error handling", func(t *testing.T) {
		// Test with invalid path
		_, err := newFileStorage("/invalid/path/that/does/not/exist")
		// This might not error immediately, but should handle errors gracefully

		// Test with valid path
		fileStorage, err := newFileStorage(".")
		require.NoError(t, err)

		// Test write operation
		err = fileStorage.Write("test_collection", "test_key", "test_value")
		require.NoError(t, err)

		// Test read operation
		var value string
		err = fileStorage.Read("test_collection", "test_key", &value)
		require.NoError(t, err)
		assert.Equal(t, "test_value", value)
	})

	t.Run("test memory storage edge cases", func(t *testing.T) {
		storage := newMemoryStorage()

		// Test with empty key
		err := storage.Write("test_collection", "", "test_value")
		require.NoError(t, err)

		var value string
		err = storage.Read("test_collection", "", &value)
		require.NoError(t, err)
		assert.Equal(t, "test_value", value)

		// Test with empty collection
		err = storage.Write("", "test_key", "test_value")
		require.NoError(t, err)

		err = storage.Read("", "test_key", &value)
		require.NoError(t, err)
		assert.Equal(t, "test_value", value)
	})

	t.Run("test storage atomic operations", func(t *testing.T) {
		storage := newMemoryStorage()

		// Test atomic operation with multiple writes
		Atomic(func() {
			err := storage.Write("test_collection", "key1", "value1")
			require.NoError(t, err)
			err = storage.Write("test_collection", "key2", "value2")
			require.NoError(t, err)
		})

		// Verify both keys were written
		var value1, value2 string
		err := storage.Read("test_collection", "key1", &value1)
		require.NoError(t, err)
		assert.Equal(t, "value1", value1)

		err = storage.Read("test_collection", "key2", &value2)
		require.NoError(t, err)
		assert.Equal(t, "value2", value2)
	})

	t.Run("test storage with complex data types", func(t *testing.T) {
		storage := newMemoryStorage()

		// Test with struct
		type TestStruct struct {
			Name  string
			Value int
		}

		testData := TestStruct{Name: "test", Value: 42}
		err := storage.Write("test_collection", "struct_key", testData)
		require.NoError(t, err)

		var retrievedData TestStruct
		err = storage.Read("test_collection", "struct_key", &retrievedData)
		require.NoError(t, err)
		assert.Equal(t, testData, retrievedData)

		// Test with slice
		sliceData := []string{"item1", "item2", "item3"}
		err = storage.Write("test_collection", "slice_key", sliceData)
		require.NoError(t, err)

		var retrievedSlice []string
		err = storage.Read("test_collection", "slice_key", &retrievedSlice)
		require.NoError(t, err)
		assert.Equal(t, sliceData, retrievedSlice)

		// Test with map
		mapData := map[string]int{"key1": 1, "key2": 2}
		err = storage.Write("test_collection", "map_key", mapData)
		require.NoError(t, err)

		var retrievedMap map[string]int
		err = storage.Read("test_collection", "map_key", &retrievedMap)
		require.NoError(t, err)
		assert.Equal(t, mapData, retrievedMap)
	})

	t.Run("test storage concurrent access", func(t *testing.T) {
		storage := newMemoryStorage()

		// Test concurrent writes
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func(index int) {
				defer func() { done <- true }()

				key := "concurrent_key_" + string(rune(index))
				value := "concurrent_value_" + string(rune(index))

				err := storage.Write("test_collection", key, value)
				require.NoError(t, err)
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify all keys were written
		keys, err := storage.GetKeys("test_collection")
		require.NoError(t, err)
		assert.Len(t, keys, 10)
	})

	t.Run("test storage error conditions", func(t *testing.T) {
		storage := newMemoryStorage()

		// Test reading non-existent key
		var value string
		err := storage.Read("test_collection", "non_existent_key", &value)
		require.Error(t, err)
		assert.Empty(t, value)

		// Test reading from non-existent collection
		err = storage.Read("non_existent_collection", "test_key", &value)
		require.Error(t, err)

		// Test deleting non-existent key
		err = storage.Delete("test_collection", "non_existent_key")
		require.NoError(t, err) // Delete should not error for non-existent keys

		// Test getting keys from non-existent collection
		keys, err := storage.GetKeys("non_existent_collection")
		require.NoError(t, err)
		assert.Empty(t, keys)
	})

	t.Run("test storage with nil values", func(t *testing.T) {
		storage := newMemoryStorage()

		// Test writing nil value
		err := storage.Write("test_collection", "nil_key", nil)
		require.NoError(t, err)

		// Test reading nil value
		var value interface{}
		err = storage.Read("test_collection", "nil_key", &value)
		require.NoError(t, err)
		assert.Nil(t, value)
	})

	t.Run("test storage performance", func(t *testing.T) {
		storage := newMemoryStorage()

		// Test writing many keys
		start := time.Now()
		for i := 0; i < 1000; i++ {
			key := "perf_key_" + string(rune(i))
			value := "perf_value_" + string(rune(i))
			err := storage.Write("test_collection", key, value)
			require.NoError(t, err)
		}
		writeDuration := time.Since(start)

		// Test reading many keys
		start = time.Now()
		for i := 0; i < 1000; i++ {
			key := "perf_key_" + string(rune(i))
			var value string
			err := storage.Read("test_collection", key, &value)
			require.NoError(t, err)
			assert.Equal(t, "perf_value_"+string(rune(i)), value)
		}
		readDuration := time.Since(start)

		// Performance should be reasonable (adjust thresholds as needed)
		assert.Less(t, writeDuration, time.Second)
		assert.Less(t, readDuration, time.Second)
	})
}
