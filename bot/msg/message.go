package msg

const typeInternal = "internal"

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
