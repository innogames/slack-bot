package queue

import (
	"strings"
	"sync"

	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/storage"
	"github.com/innogames/slack-bot/client"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	waitIcon   = "coffee"
	doneIcon   = "white_check_mark"
	storageKey = "fallback_queue"
)

var runningCommands = map[string]msg.Message{}
var mu sync.RWMutex

// AddRunningCommand registers a long running command, e.g. a running Jenkins job or watching a pull request
// it's doing following magic:
// - when using "trigger job XXX" and "then reply done" it will execute the "reply done" when the running command was finished
// - when the server got restarted, the fallbackCommand gets executed
// todo add a kill chan to kill long running command via command
// todo improve locking
func AddRunningCommand(message msg.Message, fallbackCommand string) chan<- bool {
	var queueKey string

	if fallbackCommand != "" {
		message.Text = fallbackCommand

		// add timestamp to the key to have a fix sorting by time
		queueKey = strings.ReplaceAll(message.Timestamp, ".", "") + "-" + message.GetUniqueKey()
		err := storage.Write(storageKey, queueKey, message)
		if err != nil {
			log.Error(errors.Wrap(err, "error while storing queue entry"))
		}
	}

	log.Infof("add a blocking process: %s", message.GetText())

	key := message.GetUniqueKey()

	mu.Lock()
	defer mu.Unlock()

	runningCommands[key] = message

	finished := make(chan bool, 1)

	go func() {
		defer close(finished)

		// wait until blocking task is over
		<-finished

		mu.Lock()
		delete(runningCommands, key)
		mu.Unlock()

		if queueKey != "" {
			if err := storage.Delete(storageKey, queueKey); err != nil {
				log.Error(errors.Wrapf(err, "error while deleting processed queue entry %s", key))
			}
		}
	}()

	return finished
}

// IsBlocked checks if there is a blocking command registered for this user/channel
func IsBlocked(ref msg.Ref) bool {
	_, ok := runningCommands[ref.GetUniqueKey()]

	return ok
}

// CountCurrentJobs will return the number of current pending/queued jobs
func CountCurrentJobs() int {
	mu.RLock()
	defer mu.RUnlock()

	return len(runningCommands)
}

func executeFallbackCommand() {
	keys, _ := storage.GetKeys(storageKey)

	var event msg.Message
	for _, key := range keys {
		if err := storage.Read(storageKey, key, &event); err != nil {
			log.Errorf("[Queue] Not unmarshalable: %s", err)
			continue
		}

		log.Infof("[Queue] Booted! I'll trigger this command now: `%s`", event.Text)
		client.InternalMessages <- event
	}

	storage.DeleteCollection(storageKey)
}
