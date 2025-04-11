package custom_variables

import (
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/storage"
	log "github.com/sirupsen/logrus"
)

const storeKey = "user_variables"

type list map[string]string

func loadList(userID string) list {
	userList := make(list)

	_ = storage.Read(storeKey, userID, &userList)

	return userList
}

func storeList(ref msg.Ref, list list) {
	err := storage.Write(storeKey, ref.GetUser(), list)
	if err != nil {
		log.Warnf("error while storing list: %s", err)
	}
}
