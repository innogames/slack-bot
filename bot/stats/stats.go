package stats

import (
	"sync"

	"github.com/innogames/slack-bot/v2/bot/storage"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const collection = "stats"

const (
	// TotalCommands is the tracking key to get the number of all processed commands
	TotalCommands = "command_total"

	// UnauthorizedCommands is the tracking key to get the number of commands by unauthorized users
	UnauthorizedCommands = "command_unauthorized"

	// UnknownCommands is the tracking key to get the number of all unknown commands (when the fallback-command is fired)
	UnknownCommands = "command_unknown"

	// Interactions is the tracking key to get the number of all processed interactions aka buttons
	Interactions = "interactions"
)

var mu sync.Mutex

// IncreaseOne is increasing the stats counter of the given type by 1
func IncreaseOne(key string) {
	Increase(key, 1)
}

// Increase is increasing the stats counter
func Increase(key string, count uint) {
	mu.Lock()
	defer mu.Unlock()

	var value uint
	_ = storage.Read(collection, key, &value)

	value += count

	if err := storage.Write(collection, key, value); err != nil {
		log.Warn(errors.Wrap(err, "error while increasing stats"))
	}
}

// Set the stats to a specific value
func Set(key string, value uint) {
	if err := storage.Write(collection, key, value); err != nil {
		log.Warn(errors.Wrap(err, "error while set stats"))
	}
}

// Get the counter value of of type
func Get(key string) (uint, error) {
	var value uint
	err := storage.Read(collection, key, &value)

	return value, err
}
