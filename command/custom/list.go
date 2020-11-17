package custom

import (
	"fmt"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/slack-go/slack"
)

func (c command) List(match matcher.Result, event slack.MessageEvent) {
	list := loadList(event)
	if len(list) == 0 {
		c.slackClient.Reply(event, "No commands define yet. Use `add command 'your alias' 'command to execute'`")
		return
	}

	responseText := fmt.Sprintf("You defined %d commands:", len(list))
	for alias, command := range list {
		responseText += fmt.Sprintf("\n - %s: `%s`", alias, command)
	}

	c.slackClient.Reply(event, responseText)
}
