package msg

// Message is a wrapper which holds all important fields from slack.MessageEvent
type Message struct {
	MessageRef
	Text string `json:"text,omitempty"`
}

func (msg Message) GetText() string {
	return msg.Text
}
