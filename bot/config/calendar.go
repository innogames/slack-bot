package config

import "github.com/PuloV/ics-golang"

// Calendar is able to trigger commands based on ical calendar events
type Calendar struct {
	Path   string
	Name   string
	Events []CalendarEvent
	Ical   ics.Event `yaml:"-"` // todo temporary way to pass Ical event
}

// CalendarEvent is one single calender config which should be watched
type CalendarEvent struct {
	Name     string
	Pattern  string
	Channel  string
	Commands []string
}
