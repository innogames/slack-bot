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

	// fuzzy comparison as we are working with exact time
	delta := float64(time.Millisecond * 1)

	// just started -> should return min
	actual = subject.GetNextDelay()
	assert.InDelta(t, 10, int(actual.Seconds()), delta)

	// 6h = 10 + 1/4 * 50
	subject.startedAt = time.Now().Add(-time.Hour * 6)
	actual = subject.GetNextDelay()
	assert.InDelta(t, time.Second*22+time.Millisecond*500, actual, delta)

	// > 1d -> max
	subject.startedAt = time.Now().Add(-time.Hour * 40)
	actual = subject.GetNextDelay()
	assert.InDelta(t, time.Second*60, actual, delta)
}
