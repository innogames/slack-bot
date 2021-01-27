package msg

import (
	"strconv"
	"strings"
	"time"
)

type Ref interface {
	GetChannel() string
	GetUser() string
	GetTimestamp() string
	GetThread() string
	IsInternalMessage() bool
	WithText(text string) Message
	GetUniqueKey() string
}

// MessageRef is holds meta information for an message, like author, creation date or channel
type MessageRef struct {
	Channel         string `json:"channel,omitempty"`
	User            string `json:"user,omitempty"`
	Timestamp       string `json:"ts,omitempty"`
	Thread          string `json:"thread_ts,omitempty"`
	InternalMessage bool
}

func (msg MessageRef) GetChannel() string {
	return msg.Channel
}

func (msg MessageRef) GetUser() string {
	return msg.User
}

func (msg MessageRef) GetTimestamp() string {
	return msg.Timestamp
}

func (msg MessageRef) GetThread() string {
	return msg.Thread
}

func (msg MessageRef) GetUniqueKey() string {
	return strings.TrimRight(
		strings.Join([]string{msg.GetUser(), msg.GetChannel(), msg.GetThread()}, "-"),
		"-",
	)
}

// GetTime extracts the time of the Message
func (msg MessageRef) GetTime() time.Time {
	timestamp, _ := strconv.ParseInt(msg.GetTimestamp()[0:10], 10, 64)

	return time.Unix(timestamp, 0)
}

func (msg MessageRef) IsInternalMessage() bool {
	return msg.InternalMessage
}

// WithText attaches a text to a message
func (msg MessageRef) WithText(text string) Message {
	return Message{
		MessageRef: msg,
		Text:       text,
	}
}
