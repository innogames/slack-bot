package storage

import (
	scribble "github.com/nanobox-io/golang-scribble"
	"os"
	"sync"
)

// TODO cleanup/rewrite!

var db *scribble.Driver
var dir string

var cache map[string]map[string]interface{}
var mutexes map[string]*sync.Mutex
var mutex sync.Mutex

func InitStorage(path string) (*scribble.Driver, error) {
	var err error

	cache = make(map[string]map[string]interface{})
	mutexes = make(map[string]*sync.Mutex)
	dir = path
	db, err = scribble.New(path, &scribble.Options{})

	return db, err
}

// DeleteAll will delete all stored data from the current storage
func DeleteAll() error {
	mutex.Lock()
	defer mutex.Unlock()

	cache = make(map[string]map[string]interface{})

	return os.RemoveAll(dir)
}

// Write stores one value in the persistent storage
func Write(collection string, resource string, v interface{}) error {
	lock := getOrCreateMutex(collection)
	lock.Lock()
	defer lock.Unlock()

	if _, ok := cache[collection]; !ok {
		cache[collection] = make(map[string]interface{})
	}

	cache[collection][resource] = v
	return getDb().Write(collection, resource, v)
}

// Read will load the stored data for one entry (using reference) to avoid allocation
func Read(collection string, resource string, v interface{}) error {
	lock := getOrCreateMutex(collection)
	lock.Lock()
	defer lock.Unlock()

	if _, ok := cache[collection]; !ok {
		cache[collection] = make(map[string]interface{})
	}

	// todo use fast cache lookup
	//	if value, ok := cache[collection][resource]; ok {
	//		return nil
	//	}

	err := getDb().Read(collection, resource, v)

	cache[collection][resource] = v

	return err
}

// ReadAll will return the (json) strings of a collection
func ReadAll(collection string) ([]string, error) {
	lock := getOrCreateMutex(collection)
	lock.Lock()
	defer lock.Unlock()

	return getDb().ReadAll(collection)
}

// Delete returns one entry
func Delete(collection string, resource string) error {
	lock := getOrCreateMutex(collection)
	lock.Lock()
	defer lock.Unlock()

	delete(cache[collection], resource)

	return getDb().Delete(collection, resource)
}

func getDb() *scribble.Driver {
	if db == nil {
		db, _ = InitStorage("./storage/")

		return db
	}

	return db
}

// getOrCreateMutex creates a new collection specific mutex any time a collection
// is being modified to avoid unsafe operations
func getOrCreateMutex(collection string) *sync.Mutex {
	mutex.Lock()
	defer mutex.Unlock()

	m, ok := mutexes[collection]

	// if the mutex doesn't exist make it
	if !ok {
		m = &sync.Mutex{}
		mutexes[collection] = m
	}

	return m
}
