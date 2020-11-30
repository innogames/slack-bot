package msg

const typeInternal = "internal"

// Message is a wrapper which holds all important fields from slack.MessageEvent
type Message struct {
	MessageRef
	Text string
}

func (msg Message) GetText() string {
	return msg.Text
}
