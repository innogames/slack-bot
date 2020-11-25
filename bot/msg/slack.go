package msg

import "github.com/slack-go/slack"

func (msg *Message) ToSlackEvent() slack.MessageEvent {
	subtype := ""
	if msg.InternalMessage {
		subtype = typeInternal
	}

	return slack.MessageEvent{
		Msg: slack.Msg{
			Text:            msg.Text,
			Channel:         msg.Channel,
			ThreadTimestamp: msg.Thread,
			User:            msg.User,
			Timestamp:       msg.Timestamp,
			SubType:         subtype,
		},
	}
}

func FromSlackEvent(event slack.MessageEvent) Message {
	return Message{
		Text:            event.Text,
		Channel:         event.Channel,
		Thread:          event.ThreadTimestamp,
		User:            event.User,
		Timestamp:       event.Timestamp,
		InternalMessage: event.SubType == typeInternal,
	}
}
