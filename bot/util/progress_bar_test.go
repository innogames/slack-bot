package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCalculateProgress(t *testing.T) {
	t.Run("Zero estimated duration", func(t *testing.T) {
		progress := CalculateProgress(1*time.Minute, 0)
		assert.InDelta(t, 0.0, progress, 0.001)
	})

	t.Run("Half complete", func(t *testing.T) {
		progress := CalculateProgress(30*time.Second, 60*time.Second)
		assert.InDelta(t, 0.5, progress, 0.001)
	})

	t.Run("Fully complete", func(t *testing.T) {
		progress := CalculateProgress(60*time.Second, 60*time.Second)
		assert.InDelta(t, 1.0, progress, 0.001)
	})

	t.Run("Over estimated time", func(t *testing.T) {
		progress := CalculateProgress(90*time.Second, 60*time.Second)
		assert.InDelta(t, 1.5, progress, 0.001)
	})
}

func TestRenderProgressBar(t *testing.T) {
	t.Run("Zero progress", func(t *testing.T) {
		bar := RenderProgressBar(0.0)
		assert.Equal(t, "░░░░░░░░░░░░░░░░░░░░ 0%", bar)
	})

	t.Run("25% progress", func(t *testing.T) {
		bar := RenderProgressBar(0.25)
		assert.Equal(t, "█████░░░░░░░░░░░░░░░ 25%", bar)
	})

	t.Run("50% progress", func(t *testing.T) {
		bar := RenderProgressBar(0.5)
		assert.Equal(t, "██████████░░░░░░░░░░ 50%", bar)
	})

	t.Run("75% progress", func(t *testing.T) {
		bar := RenderProgressBar(0.75)
		assert.Equal(t, "███████████████░░░░░ 75%", bar)
	})

	t.Run("100% progress", func(t *testing.T) {
		bar := RenderProgressBar(1.0)
		assert.Equal(t, "████████████████████ 100%", bar)
	})

	t.Run("Over 100% progress", func(t *testing.T) {
		bar := RenderProgressBar(1.5)
		assert.Equal(t, "████████████████████ 150%", bar)
	})

	t.Run("Negative progress", func(t *testing.T) {
		bar := RenderProgressBar(-0.1)
		assert.Equal(t, "░░░░░░░░░░░░░░░░░░░░ -10%", bar)
	})
}

func TestRenderCountProgressBar(t *testing.T) {
	t.Run("No jobs", func(t *testing.T) {
		assert.Empty(t, RenderCountProgressBar(0, 0))
	})

	t.Run("Half done", func(t *testing.T) {
		bar := RenderCountProgressBar(5, 10)
		assert.Equal(t, "██████████░░░░░░░░░░ 50% (5/10)", bar)
	})

	t.Run("All done", func(t *testing.T) {
		bar := RenderCountProgressBar(8, 8)
		assert.Equal(t, "████████████████████ 100% (8/8)", bar)
	})
}
