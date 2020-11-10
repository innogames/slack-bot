package util

import (
	"time"
)

// after this duration, return the maxDuration. Below, calculate a linear delay, based on min/max
const durationForMaxDelay = time.Hour * 24

// Returns a increasing duration to have less polling overhead but keep a higher frequency in the first day
//
// given: min: 2, max: 9
// Timeline: start                       max             now
// Result:   22223333444455556666777788889999999999999999999
func GetIncreasingDelay(minDuration time.Duration, maxDuration time.Duration) IncreasingDelay {
	return IncreasingDelay{
		time.Now(),
		minDuration,
		maxDuration,
	}
}

type IncreasingDelay struct {
	startedAt   time.Time
	minDuration time.Duration
	maxDuration time.Duration
}

func (d IncreasingDelay) GetNextDelay() time.Duration {
	startedAgo := time.Since(d.startedAt)

	if startedAgo >= durationForMaxDelay {
		return d.maxDuration
	}

	additional := float64((d.maxDuration - d.minDuration).Nanoseconds()) * (startedAgo.Minutes() / durationForMaxDelay.Minutes())

	return d.minDuration + time.Duration(additional)
}
