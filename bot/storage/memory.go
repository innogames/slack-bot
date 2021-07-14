package storage

import (
	"encoding/json"
	"fmt"
	"sync"
)

type memoryCollection map[string][]byte

func newMemoryStorage() Storage {
	return &memoryStorage{
		storage: make(map[string]memoryCollection),
		mutex:   sync.RWMutex{},
	}
}

type memoryStorage struct {
	storage map[string]memoryCollection
	mutex   sync.RWMutex
}

func (s *memoryStorage) Write(collection, key string, v interface{}) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

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

func (s *memoryStorage) Read(collection, key string, v interface{}) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if _, ok := s.storage[collection]; !ok {
		return fmt.Errorf("collection is empty")
	}

	if _, ok := s.storage[collection][key]; !ok {
		return fmt.Errorf("value is empty")
	}

	return json.Unmarshal(s.storage[collection][key], v)
}

func (s *memoryStorage) GetKeys(collection string) ([]string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	keys := make([]string, 0, len(s.storage[collection]))

	for key := range s.storage[collection] {
		keys = append(keys, key)
	}

	return keys, nil
}

func (s *memoryStorage) Delete(collection, key string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.storage[collection], key)

	return nil
}
