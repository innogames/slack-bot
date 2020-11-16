package util

import (
	"regexp"
	"strings"
	"time"
)

var stripDecimalPlace = regexp.MustCompile(`(\d+)\.\d+([µa-z]+)`)
var durationReplacer = strings.NewReplacer(
	"min", "m",
	"sec", "s",
)

// ParseDuration also allowes other duration modifier like "min" or "sec"
// e.g. 12min10sec -> 12m10s
func ParseDuration(input string) (time.Duration, error) {
	input = durationReplacer.Replace(input)

	return time.ParseDuration(input)
}

// FormatDuration shortens a duration string representation.
// e.g. "12m1.231001s" -> "12m1s"
func FormatDuration(duration time.Duration) string {
	output := duration.String()

	return stripDecimalPlace.ReplaceAllString(output, "$1$2")
}
