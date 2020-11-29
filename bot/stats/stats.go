package stats

import (
	"github.com/innogames/slack-bot/bot/storage"
	"sync"
)

const collection = "stats"

const (
	// keys of some tracked statistics
	TotalCommands        = "command_total"
	UnauthorizedCommands = "command_unauthorized"
	UnknownCommands      = "command_unknown"
)

var mu sync.Mutex

// IncreaseOne is increasing the stats counter of the given type by 1
func IncreaseOne(key string) {
	Increase(key, 1)
}

// IncreaseOne is increasing the stats counter
func Increase(key string, count uint) {
	mu.Lock()
	defer mu.Unlock()

	var value uint
	storage.Read(collection, key, &value)

	value += count

	storage.Write(collection, key, value)
}

// Get the counter value of of type
func Get(key string) (uint, error) {
	var value uint
	err := storage.Read(collection, key, &value)

	return value, err
}
