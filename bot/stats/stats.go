package stats

import (
	"github.com/innogames/slack-bot/bot/storage"
	"sync"
)

const collection = "stats"

const (
	TotalCommands        = "command_total"
	UnauthorizedCommands = "command_unauthorized"
	UnknownCommands      = "command_unknown"
)

var mu sync.Mutex

func IncreaseOne(key string) {
	Increase(key, 1)
}

func Increase(key string, count uint) {
	mu.Lock()
	defer mu.Unlock()

	var value uint
	storage.Read(collection, key, &value)

	value += count

	storage.Write(collection, key, value)
}

func Get(key string) (uint, error) {
	var value uint
	err := storage.Read(collection, key, &value)

	return value, err
}
