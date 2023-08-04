package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	list := []string{"foo", "bar", "baz"}
	assert.False(t, Contains(list, ""))
	assert.False(t, Contains(list, "XXX"))
	assert.True(t, Contains(list, "foo"))

	list2 := []int{1, 2, 3}
	assert.False(t, Contains(list2, -1))
	assert.False(t, Contains(list2, 0))
	assert.True(t, Contains(list2, 2))
}
