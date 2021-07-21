package msg

// Message is a wrapper which holds all important fields from slack.MessageEvent
import "sync"

// Message represents a slack.Message in a slim format. The MessageRef contains the context of the message
type Message struct {
	MessageRef
	Text string          `json:"text,omitempty"`
	Done *sync.WaitGroup `json:"-"` // WaitGroup gets unlocked when the message was processed
}

// GetText returns the attached text of the message
func (msg *Message) GetText() string {
	return msg.Text
}

// AddDoneHandler will register a sync.WaitGroup
func (msg *Message) AddDoneHandler() *sync.WaitGroup {
	msg.Done = &sync.WaitGroup{}
	msg.Done.Add(1)

	return msg.Done
}
