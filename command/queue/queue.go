package queue

import (
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/storage"
	"github.com/innogames/slack-bot/client"
	log "github.com/sirupsen/logrus"
	"strings"
	"sync"
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
func AddRunningCommand(message msg.Message, fallbackCommand string) chan bool {
	var queueKey string

	if fallbackCommand != "" {
		message.Text = fallbackCommand

		// add timestamp to the key to have a fix sorting by time
		queueKey = strings.ReplaceAll(message.Timestamp, ".", "") + "-" + message.GetUniqueKey()
		storage.Write(storageKey, queueKey, message)
	}

	log.Infof("add a blocking process: %s", message.GetText())

	key := getKey(message)

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
			storage.Delete(storageKey, queueKey)
		}
	}()

	return finished
}

// IsBlocked checks if there is a blocking command registered for this user/channel
func IsBlocked(ref msg.Ref) bool {
	_, ok := runningCommands[getKey(ref)]

	return ok
}

// CountCurrentJobs will return the number of current pending/queued jobs
func CountCurrentJobs() int {
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

func getKey(ref msg.Ref) string {
	return ref.GetUser() + ref.GetChannel()
}
