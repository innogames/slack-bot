package util

import (
	"math/rand"
	"time"
)

// after this duration, return the maxDuration. Below, calculate a linear delay, based on min/max
const durationForMaxDelay = time.Hour * 24

// GetIncreasingDelay Returns a increasing duration to have less polling overhead but keep a higher frequency in the first day
//
// given: min: 2, max: 9
// Timeline: start                       max             now
// Result:   22223333444455556666777788889999999999999999999
func GetIncreasingDelay(minDuration time.Duration, maxDuration time.Duration) IncreasingDelay {
	return IncreasingDelay{
		time.Millisecond * time.Duration(rand.Intn(2000)), //nolint:gosec // add up to 2s randomly to avoid peaks
		time.Now(),
		minDuration,
		maxDuration,
	}
}

// IncreasingDelay is a wrapper to support GetIncreasingDelay to have a increasing interval functionality
type IncreasingDelay struct {
	randomAdd   time.Duration
	startedAt   time.Time
	minDuration time.Duration
	maxDuration time.Duration
}

// GetNextDelay returns a time.Duration based on the given context
func (d IncreasingDelay) GetNextDelay() time.Duration {
	startedAgo := time.Since(d.startedAt)

	if startedAgo >= durationForMaxDelay {
		return d.maxDuration
	}

	additional := float64((d.maxDuration - d.minDuration).Nanoseconds()) * (startedAgo.Minutes() / durationForMaxDelay.Minutes())

	return d.minDuration + time.Duration(additional) + d.randomAdd
}
