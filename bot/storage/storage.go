package storage

import (
	"fmt"
	"regexp"
	"sync"
)

var (
	currentStorage Storage
	mu             sync.Mutex
)

// Storage is the main interface which is used to persist bot related data (like queued messages or user histories)
type Storage interface {
	Write(collection, key string, v interface{}) error
	Read(collection, key string, v interface{}) error
	GetKeys(collection string) ([]string, error)
	Delete(collection, key string) error
}

// allowed characters for stage keys/collection
var keyRegexp = regexp.MustCompile(`^[\w\-,+@]+$`)

// InitStorage registers a local directory as JSON file Storage
func InitStorage(path string) error {
	var err error

	if path == "" {
		currentStorage = newMemoryStorage()
	} else {
		currentStorage, err = newFileStorage(path)
	}

	return err
}

// SetStorage provide Storage to persist data for bot usage
func SetStorage(storage Storage) {
	currentStorage = storage
}

// Write stores one value in the persistent Storage
func Write(collection string, key string, v interface{}) error {
	if err := validateKey(collection, key); err != nil {
		return err
	}

	return getStorage().Write(collection, key, v)
}

// Read will load the stored data for one entry (using reference) to avoid allocation
func Read(collection string, key string, v interface{}) error {
	if err := validateKey(collection, key); err != nil {
		return err
	}

	return getStorage().Read(collection, key, v)
}

// GetKeys will return the (json) strings of a collection
func GetKeys(collection string) ([]string, error) {
	if err := validateKey(collection); err != nil {
		return nil, err
	}

	return getStorage().GetKeys(collection)
}

// DeleteCollection delete all entries of a collection
func DeleteCollection(collection string) error {
	if err := validateKey(collection); err != nil {
		return err
	}

	return getStorage().Delete(collection, "")
}

// Delete will return a single entry of a collection
func Delete(collection string, key string) error {
	if err := validateKey(collection, key); err != nil {
		return err
	}

	return getStorage().Delete(collection, key)
}

// check if a given key/collection only contains a subset of valid characters
func validateKey(keys ...string) error {
	for _, key := range keys {
		if !keyRegexp.MatchString(key) {
			return fmt.Errorf("invalid Storage key: %s", key)
		}
	}

	return nil
}

func getStorage() Storage {
	if currentStorage != nil {
		return currentStorage
	}

	mu.Lock()
	defer mu.Unlock()

	currentStorage = newMemoryStorage()

	return currentStorage
}
