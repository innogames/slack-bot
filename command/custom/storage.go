package custom

import (
	"github.com/innogames/slack-bot/bot/storage"
	"github.com/nlopes/slack"
)

const storeKey = "user_commands"

type list map[string]string

func loadList(event slack.MessageEvent) list {
	list := make(list, 0)

	storage.Read(storeKey, event.User, &list)

	return list
}

func storeList(event slack.MessageEvent, list list) {
	storage.Write(storeKey, event.User, list)
}
