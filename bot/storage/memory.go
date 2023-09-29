package storage

import (
	"encoding/json"
	"fmt"

	"github.com/innogames/slack-bot/v2/bot/util"
	"golang.org/x/exp/maps"
)

// this is a primitive in-memory storage which is used for faster storage of data.
// we use JSON serialization here which sounds quite inefficient, but we want the same behavior as other storages:
// - immutable data storage
// - same behavior for json tags
type memoryCollection map[string][]byte

func newMemoryStorage() Storage {
	return &memoryStorage{
		storage: make(map[string]memoryCollection),
		locks:   util.NewGroupedLogger(),
	}
}

type memoryStorage struct {
	storage map[string]memoryCollection
	locks   util.GroupedLock[string]
}

func (s *memoryStorage) Write(collection, key string, v any) error {
	lock := s.locks.GetLock(collection)
	defer lock.Unlock()

	if _, ok := s.storage[collection]; !ok {
		s.storage[collection] = make(memoryCollection)
	}

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	s.storage[collection][key] = data

	return nil
}

func (s *memoryStorage) Read(collection, key string, v any) error {
	lock := s.locks.GetRLock(collection)
	defer lock.Unlock()

	if _, ok := s.storage[collection]; !ok {
		return fmt.Errorf("collection is empty")
	}

	if _, ok := s.storage[collection][key]; !ok {
		return fmt.Errorf("value is empty")
	}

	return json.Unmarshal(s.storage[collection][key], v)
}

func (s *memoryStorage) GetKeys(collection string) ([]string, error) {
	lock := s.locks.GetRLock(collection)
	defer lock.Unlock()

	keys := maps.Keys(s.storage[collection])

	return keys, nil
}

func (s *memoryStorage) Delete(collection, key string) error {
	lock := s.locks.GetLock(collection)
	defer lock.Unlock()

	delete(s.storage[collection], key)

	return nil
}
