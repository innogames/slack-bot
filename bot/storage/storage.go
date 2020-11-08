package storage

import "sync"

var currentStorage storage
var mu sync.Mutex

type storage interface {
	Write(collection, key string, v interface{}) error
	Read(collection, key string, v interface{}) error
	GetKeys(collection string) ([]string, error)
	Delete(collection, key string) error
}

// InitStorage registers a local directory as JSON file storage
func InitStorage(path string) error {
	var err error
	if path == "" {
		currentStorage = newMemoryStorage()
	} else {
		currentStorage, err = newFileStorage(path)
	}

	return err
}

// SetStorage provide storage to persist data for bot usage
func SetStorage(storage storage) {
	currentStorage = storage
}

// Write stores one value in the persistent storage
func Write(collection string, key string, v interface{}) error {
	return getStorage().Write(collection, key, v)
}

// Read will load the stored data for one entry (using reference) to avoid allocation
func Read(collection string, key string, v interface{}) error {
	return getStorage().Read(collection, key, v)
}

// GetKeys will return the (json) strings of a collection
func GetKeys(collection string) ([]string, error) {
	return getStorage().GetKeys(collection)
}

// Delete returns one entry
func Delete(collection string, key string) error {
	return getStorage().Delete(collection, key)
}

func getStorage() storage {
	mu.Lock()
	defer mu.Unlock()

	if currentStorage == nil {
		currentStorage = newMemoryStorage()
	}

	return currentStorage
}
