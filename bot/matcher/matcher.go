package matcher

import (
	"github.com/innogames/slack-bot/bot/msg"
)

// Runner contains the actual logic of the executed command. gets the actual slack MessageEvent and the matched parameters of the message
type Runner func(match Result, message msg.Message)

// Matcher is executed for each command and checks if it should execute the Runner for the given event text
type Matcher interface {
	Match(message msg.Message) (Runner, Result)
}
