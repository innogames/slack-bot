package custom_commmands

import (
	"fmt"
	"strings"

	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/client"
)

// check if the current user has a alias for the current message, if yes, execute the commands
func (c command) handle(ref msg.Ref, text string) bool {
	var commands string

	list := loadList(ref)
	if commands = list[text]; commands == "" {
		return false
	}

	c.SendEphemeralMessage(ref, fmt.Sprintf("executing command: `%s`", commands))
	for command := range strings.SplitSeq(commands, ";") {
		message := client.HandleMessageWithDoneHandler(ref.WithText(command))
		message.Wait()
	}

	return true
}
