package util

import (
	"strconv"
	"strings"
	"time"

	"github.com/nlopes/slack"
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

// GetMessageTime will return a time.Time of the sent date from the MessageEvent
func GetMessageTime(event slack.MessageEvent) time.Time {
	timestamp, _ := strconv.ParseInt(event.Timestamp[0:10], 10, 64)

	return time.Unix(timestamp, 0)
}
