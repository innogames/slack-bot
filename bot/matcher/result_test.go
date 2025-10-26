package matcher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResultEdgeCases(t *testing.T) {
	t.Run("invalid integer parsing", func(t *testing.T) {
		result := Result{
			"valid_number":   "123",
			"invalid_number": "abc",
			"empty_string":   "",
			"zero_string":    "0",
		}

		assert.Equal(t, 123, result.GetInt("valid_number"))
		assert.Equal(t, 0, result.GetInt("invalid_number")) // Should return 0 for invalid
		assert.Equal(t, 0, result.GetInt("empty_string"))   // Should return 0 for empty
		assert.Equal(t, 0, result.GetInt("zero_string"))    // Should return 0 for "0"
		assert.Equal(t, 0, result.GetInt("nonexistent"))    // Should return 0 for missing
	})

	t.Run("has logic variations", func(t *testing.T) {
		result := Result{
			"present": "value",
			"empty":   "",
			"false":   "false",
			"zero":    "0",
		}

		assert.True(t, result.Has("present"))
		assert.False(t, result.Has("empty"))
		assert.False(t, result.Has("false"))
		assert.True(t, result.Has("zero")) // "0" is considered present
		assert.False(t, result.Has("nonexistent"))
	})

	t.Run("string retrieval", func(t *testing.T) {
		result := Result{
			"key1": "value1",
			"key2": "",
		}

		assert.Equal(t, "value1", result.GetString("key1"))
		assert.Equal(t, "", result.GetString("key2"))
		assert.Equal(t, "", result.GetString("nonexistent"))
	})
}
