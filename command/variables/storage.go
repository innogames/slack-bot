package variables

import (
	"github.com/innogames/slack-bot/bot/storage"
	"github.com/slack-go/slack"
)

const storeKey = "user_variables"

type list map[string]string

func loadList(userID string) list {
	list := make(list)

	storage.Read(storeKey, userID, &list)

	return list
}

func storeList(event slack.MessageEvent, list list) {
	storage.Write(storeKey, event.User, list)
}
