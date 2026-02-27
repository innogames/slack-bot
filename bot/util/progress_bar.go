package util

import (
	"fmt"
	"strings"
	"time"
)

// CalculateProgress returns a progress percentage (0.0 to 1.0+)
func CalculateProgress(elapsed, estimated time.Duration) float64 {
	if estimated <= 0 {
		return 0
	}
	return float64(elapsed) / float64(estimated)
}

// RenderProgressBar creates a visual progress bar using Unicode blocks
func RenderProgressBar(progress float64) string {
	const barLength = 20
	filled := max(
		min(int(progress*float64(barLength)), barLength),
		0,
	)

	var bar strings.Builder
	for i := range barLength {
		if i < filled {
			bar.WriteString("█")
		} else {
			bar.WriteString("░")
		}
	}

	percentage := int(progress * 100)
	return fmt.Sprintf("%s %d%%", bar.String(), percentage)
}

// RenderCountProgressBar creates a progress bar based on completed/total counts
func RenderCountProgressBar(done, total int) string {
	if total == 0 {
		return ""
	}
	progress := float64(done) / float64(total)
	return fmt.Sprintf("%s (%d/%d)", RenderProgressBar(progress), done, total)
}
