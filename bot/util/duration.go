package util

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	daysRe            = regexp.MustCompile(`^(\d+)d$`)
	stripDecimalPlace = regexp.MustCompile(`(\d+)\.(\d)\d*([µa-z]+)`)
	durationReplacer  = strings.NewReplacer(
		"min", "m",
		"sec", "s",
		"days", "d",
		"day", "d",
	)
)

// ParseDuration also allows other duration modifier like "min" or "sec"
// e.g. 12min10sec -> 12m10s
func ParseDuration(input string) (time.Duration, error) {
	input = durationReplacer.Replace(input)

	if match := daysRe.FindStringSubmatch(input); len(match) > 0 {
		days, _ := strconv.Atoi(match[1])

		return time.Hour * 24 * time.Duration(days), nil
	}

	return time.ParseDuration(input)
}

// FormatDuration shortens a duration string representation.
// e.g. "12m1.231001s" -> "12m1s"
func FormatDuration(duration time.Duration) string {
	output := ""

	// extract days out of duration
	fullDays := int(duration.Hours() / 24)
	if fullDays > 0 {
		duration -= time.Hour * 24 * time.Duration(fullDays)
		output += fmt.Sprintf("%dd", fullDays)
	}
	output += duration.String()

	return stripDecimalPlace.ReplaceAllString(output, "$1.$2$3")
}
