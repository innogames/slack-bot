package msg

import "github.com/slack-go/slack"

// GetMessageRef create a msg.ItemRef, an unique identifier to a message
func (msg *Message) GetMessageRef() slack.ItemRef {
	return slack.NewRefToMessage(msg.Channel, msg.Timestamp)
}

func FromSlackEvent(event slack.MessageEvent) Message {
	return Message{
		Text: event.Text,
		MessageRef: MessageRef{
			Channel:         event.Channel,
			Thread:          event.ThreadTimestamp,
			User:            event.User,
			Timestamp:       event.Timestamp,
			InternalMessage: event.SubType == typeInternal,
		},
	}
}
