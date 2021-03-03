package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIncreasingDelay(t *testing.T) {
	subject := GetIncreasingDelay(time.Second*10, time.Second*60)
	subject.randomAdd = time.Duration(0)
	var actual time.Duration

	// just started -> should return min
	actual = subject.GetNextDelay()
	assert.Equal(t, 10, int(actual.Seconds()))

	// 6h = 10 + 1/4 * 50
	subject.startedAt = time.Now().Add(-time.Hour * 6)
	actual = subject.GetNextDelay()
	assert.Equal(t, time.Second*22+time.Millisecond*500, actual)

	// > 1d -> max
	subject.startedAt = time.Now().Add(-time.Hour * 40)
	actual = subject.GetNextDelay()
	assert.Equal(t, time.Second*60, actual)
}
