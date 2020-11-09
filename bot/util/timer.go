package util

import (
	"fmt"
	"time"
)

func GetIncreasingTime(defaultDuration time.Duration, counter int) time.Duration {
	fmt.Printf("%s, %d\n", defaultDuration, counter)

	return defaultDuration

	//additionalMs := defaultDuration.Milliseconds() * (counter * 1.1)
	//return defaultDuration + (time.Millisecond * additionalMs)
}
