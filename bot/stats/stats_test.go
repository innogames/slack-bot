package stats

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStats(t *testing.T) {
	value, err := Get("test")
	assert.NotNil(t, err)
	assert.Equal(t, value, uint(0))

	Increase("test", 2)
	value, err = Get("test")
	assert.Nil(t, err)
	assert.Equal(t, value, uint(2))

	IncreaseOne("test")
	value, err = Get("test")
	assert.Nil(t, err)
	assert.Equal(t, value, uint(3))

	Set("test", 42)
	value, err = Get("test")
	assert.Nil(t, err)
	assert.Equal(t, value, uint(42))
}
