package msg

import (
	"strconv"
	"strings"
	"time"
)

// Ref is the context of a message: the author, channel, timestamp etc
type Ref interface {
	GetChannel() string
	GetUser() string
	GetTimestamp() string
	GetThread() string
	IsInternalMessage() bool
	IsUpdatedMessage() bool
	WithText(text string) Message
	GetUniqueKey() string
}

// MessageRef is holds meta information for an message, like author, creation date or channel
type MessageRef struct {
	Channel         string `json:"channel,omitempty"`
	User            string `json:"user,omitempty"`
	Timestamp       string `json:"ts,omitempty"`
	Thread          string `json:"thread_ts,omitempty"`
	InternalMessage bool   `json:"InternalMessage,omitempty"`
	UpdatedMessage  bool   `json:"updated,omitempty"`
}

// GetChannel returns the channel id (usually starting with "C") of the current message
func (msg MessageRef) GetChannel() string {
	return msg.Channel
}

// GetUser returns the user id (usually starting with "U") of the current message
func (msg MessageRef) GetUser() string {
	return msg.User
}

// GetTimestamp returns the slack-timestamp of the message (e.g. 1613728332.201900)
func (msg MessageRef) GetTimestamp() string {
	return msg.Timestamp
}

// GetThread returns the thread "timestamp" of the message
func (msg MessageRef) GetThread() string {
	return msg.Thread
}

// IsUpdatedMessage identifies if the processed message should be updated (like with a "refresh" button)
func (msg MessageRef) IsUpdatedMessage() bool {
	return msg.UpdatedMessage
}

// GetUniqueKey generated a unique identifier for a message (based on the user/channel/thread)
func (msg MessageRef) GetUniqueKey() string {
	key := strings.TrimRight(
		strings.Join([]string{msg.GetUser(), msg.GetChannel(), msg.GetThread()}, "-"),
		"-",
	)
	return strings.ReplaceAll(key, ".", "_")
}

// GetTime extracts the time.Time of the Message
func (msg MessageRef) GetTime() time.Time {
	if msg.GetTimestamp() == "" {
		return time.Now()
	}

	// looks like 1628614631.250000
	parts := strings.SplitN(msg.GetTimestamp(), ".", 2)

	timestamp, _ := strconv.ParseInt(parts[0], 10, 64)
	micro, _ := strconv.ParseInt(parts[1], 10, 64)

	return time.Unix(timestamp, micro*1000)
}

// IsInternalMessage is set when the bot is generating internal messages which are handles (like from "cron" command)
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
