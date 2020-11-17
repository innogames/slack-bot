package stats

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStats(t *testing.T) {
	value, err := Get("test")
	assert.NotNil(t, err)
	assert.Equal(t, value, 0)

	Increase("test", 2)
	value, err = Get("test")
	assert.Nil(t, err)
	assert.Equal(t, value, 2)

	IncreaseOne("test")
	value, err = Get("test")
	assert.Nil(t, err)
	assert.Equal(t, value, 3)
}
