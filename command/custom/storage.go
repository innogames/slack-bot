package custom

import (
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/storage"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const storeKey = "user_commands"

type list map[string]string

func loadList(ref msg.Ref) list {
	list := make(list)

	storage.Read(storeKey, ref.GetUser(), &list)

	return list
}

func storeList(ref msg.Ref, list list) {
	err := storage.Write(storeKey, ref.GetUser(), list)
	if err != nil {
		log.Error(errors.Wrap(err, "error wile storing custom command"))
	}
}
