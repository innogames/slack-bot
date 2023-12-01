package stats

import (
	"github.com/innogames/slack-bot/v2/bot/storage"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/constraints"
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

// IncreaseOne is increasing the stats counter of the given type by 1
func IncreaseOne(key string) {
	Increase(key, 1)
}

// Increase is increasing the stats counter
func Increase[T constraints.Signed](key string, count T) {
	storage.Atomic(func() {
		var value T
		_ = storage.Read(collection, key, &value)

		value += count

		Set(key, value)
	})
}

// Set the stats to a specific value
func Set[T constraints.Signed](key string, value T) {
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

// GetKeys returns all used metric keys
func GetKeys() []string {
	keys, _ := storage.GetKeys(collection)

	return keys
}
