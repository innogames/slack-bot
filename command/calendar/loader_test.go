package calendar

import (
	"github.com/innogames/slack-bot/bot/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCalendar(t *testing.T) {
	t.Run("Loader", func(t *testing.T) {
		calendar := loadCalender(config.Calendar{
			Name: "Test",
			Path: "test.ics",
		})

		events := calendar.GetEvents()

		assert.Equal(t, 2, len(events))

		assert.Equal(t, "General Operative Meeting", events[0].GetSummary())
		assert.Equal(t, "Geometry Exam", events[1].GetSummary())
	})
}
