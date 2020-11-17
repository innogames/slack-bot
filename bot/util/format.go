package util

import "fmt"

func FormatBytes(byte uint64) string {
	const unit = 1000
	if byte < unit {
		return fmt.Sprintf("%d B", byte)
	}

	div, exp := int64(unit), 0
	for n := byte / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf(
		"%.1f %cB",
		float64(byte)/float64(div), "kMGTPE"[exp],
	)
}
