package stats

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStats(t *testing.T) {
	value, err := Get("test")
	require.Error(t, err)
	assert.Equal(t, uint(0), value)

	Increase("test", 2)
	value, err = Get("test")
	require.NoError(t, err)
	assert.Equal(t, uint(2), value)

	IncreaseOne("test")
	value, err = Get("test")
	require.NoError(t, err)
	assert.Equal(t, uint(3), value)

	Set("test", 42)
	value, err = Get("test")
	require.NoError(t, err)
	assert.Equal(t, uint(42), value)

	Increase("test", int64(1))
	Increase("test", int8(1))
	Increase("test", 1)
	value, err = Get("test")
	require.NoError(t, err)
	assert.Equal(t, uint(45), value)

	IncreaseOne("test")
	value, err = Get("test")
	require.NoError(t, err)
	assert.Equal(t, uint(46), value)
}
