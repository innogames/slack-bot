package msg

// Message is a wrapper which holds all important fields from slack.MessageEvent
import "sync"

type Message struct {
	MessageRef
	Done *sync.WaitGroup
	Text string `json:"text,omitempty"`
}

func (msg *Message) GetText() string {
	return msg.Text
}

// AddDoneHandler will register a sync.WaitGroup
func (msg *Message) AddDoneHandler() *sync.WaitGroup {
	msg.Done = &sync.WaitGroup{}
	msg.Done.Add(1)

	return msg.Done
}
