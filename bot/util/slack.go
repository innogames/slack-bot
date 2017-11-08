package util

import (
	"github.com/nlopes/slack"
	"strings"
)

// GetFullEventKey builds a key over user-channel-threadtimestamp
func GetFullEventKey(event slack.MessageEvent) string {
	return strings.TrimRight(
		strings.Join([]string{event.User, event.Channel, event.ThreadTimestamp}, "-"),
		"-",
	)
}

// GetShortEventKey builds a key over user-channel
func GetShortEventKey(event slack.MessageEvent) string {
	return strings.TrimRight(
		strings.Join([]string{event.User, event.Channel}, "-"),
		"-",
	)
}
