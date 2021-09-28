package queue

import (
	"strings"
	"sync"

	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/storage"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	waitIcon   = "coffee"
	doneIcon   = "white_check_mark"
	storageKey = "fallback_queue"
)

var mu sync.RWMutex

// AddRunningCommand registers a long running command, e.g. a running Jenkins job or watching a pull request
// it's doing following magic:
// - when using "trigger job XXX" and "then reply done" it will execute the "reply done" when the running command was finished
// - when the server got restarted, the fallbackCommand gets executed
func AddRunningCommand(message msg.Message, fallbackCommand string) *RunningCommand {
	var queueKey string

	// store fallback command in storage:
	// when the bot restarts for any reason, it can recover, based on this fallback commands
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

	mu.Lock()
	defer mu.Unlock()

	key := message.GetUniqueKey()

	runningCommand := &RunningCommand{}
	runningCommand.wg.Add(1)
	runningCommands[key] = runningCommand

	go func() {
		// wait until blocking task is over
		runningCommand.Wait()

		mu.Lock()
		delete(runningCommands, key)
		mu.Unlock()

		if queueKey != "" {
			if err := storage.Delete(storageKey, queueKey); err != nil {
				log.Error(errors.Wrapf(err, "error while deleting processed queue entry %s", key))
			}
		}
	}()

	return runningCommand
}

// CountCurrentJobs will return the number of current pending/queued jobs
func CountCurrentJobs() int {
	mu.RLock()
	defer mu.RUnlock()

	return len(runningCommands)
}

func executeFallbackCommand() {
	keys, _ := storage.GetKeys(storageKey)
	if len(keys) == 0 {
		return
	}

	log.Infof("[Queue] Booted! I'll trigger %d command now", len(keys))

	var event msg.Message
	for _, key := range keys {
		if err := storage.Read(storageKey, key, &event); err != nil {
			log.Errorf("[Queue] Not unmarshalable: %s", err)
			continue
		}

		client.HandleMessage(event)
	}

	_ = storage.DeleteCollection(storageKey)
}
