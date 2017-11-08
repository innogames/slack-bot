package util

import (
	"regexp"
	"strings"
	"time"
)

var stripDecimalPlace = regexp.MustCompile(`(\d+)\.\d+([µa-z]+)`)

// ParseDuration also allowes other duration modifier like "min" or "sec"
// e.g. 12min10sec -> 12m10s
func ParseDuration(input string) (time.Duration, error) {
	input = strings.Replace(input, "min", "m", 1)
	input = strings.Replace(input, "sec", "s", 1)

	return time.ParseDuration(input)
}

// FormatDuration shortens a duration string representation.
// e.g. "12m1.231001s" -> "12m1s"
func FormatDuration(duration time.Duration) string {
	output := duration.String()

	return stripDecimalPlace.ReplaceAllString(output, "$1$2")
}
