package variables

import (
	"github.com/innogames/slack-bot/bot/storage"
	"github.com/nlopes/slack"
)

const storeKey = "user_variables"

type list map[string]string

func loadList(userId string) list {
	list := make(list, 0)

	storage.Read(storeKey, userId, &list)

	return list
}

func storeList(event slack.MessageEvent, list list) {
	storage.Write(storeKey, event.User, list)
}
