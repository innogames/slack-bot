package custom

import (
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/storage"
)

const storeKey = "user_commands"

type list map[string]string

func loadList(ref msg.Ref) list {
	list := make(list)

	storage.Read(storeKey, ref.GetUser(), &list)

	return list
}

func storeList(ref msg.Ref, list list) {
	storage.Write(storeKey, ref.GetUser(), list)
}
