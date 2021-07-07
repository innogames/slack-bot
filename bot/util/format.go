package util

import (
	"fmt"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// FormatBytes formats a given number of bytes in a simple human readable version
func FormatBytes(bytes uint64) string {
	const unit = 1000
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf(
		"%.1f %cB",
		float64(bytes)/float64(div), "kMGTPE"[exp],
	)
}

// FormatInt returns the given integer in a english representation (like 123,456,789)
func FormatInt(number int) string {
	p := message.NewPrinter(language.English)

	return p.Sprintf("%d", number)
}
