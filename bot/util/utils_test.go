package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUtils(t *testing.T) {
	t.Run("RandomString", func(t *testing.T) {
		actual := RandString(1)
		assert.Len(t, actual, 1)

		actual = RandString(10)
		assert.Len(t, actual, 10)
	})
}
