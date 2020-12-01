package variables

import (
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/storage"
)

const storeKey = "user_variables"

type list map[string]string

func loadList(userID string) list {
	list := make(list)

	storage.Read(storeKey, userID, &list)

	return list
}

func storeList(ref msg.Ref, list list) {
	storage.Write(storeKey, ref.GetUser(), list)
}
