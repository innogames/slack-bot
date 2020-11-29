package msg

const typeInternal = "internal"

// Message is a wrapper which holds all important fields from slack.MessageEvent
type Message struct {
	Text            string
	Channel         string
	Thread          string
	User            string
	Timestamp       string
	InternalMessage bool
}

func (msg *Message) IsInternalMessage() bool {
	return msg.InternalMessage
}
