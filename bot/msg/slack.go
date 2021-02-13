package msg

import "github.com/slack-go/slack"

// GetMessageRef create a msg.ItemRef, an unique identifier to a message
func (msg *Message) GetMessageRef() slack.ItemRef {
	return slack.NewRefToMessage(msg.Channel, msg.Timestamp)
}

// FromSlackEvent generates a slack.MessageEvent into a msg.Message
func FromSlackEvent(event *slack.MessageEvent) Message {
	return Message{
		Text: event.Text,
		MessageRef: MessageRef{
			Channel:   event.Channel,
			Thread:    event.ThreadTimestamp,
			User:      event.User,
			Timestamp: event.Timestamp,
		},
	}
}
