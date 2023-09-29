package storage

// NewChainStorage combines two storages to have a persistent and fast a memory storage
func NewChainStorage(persistent Storage, memory Storage) Storage {
	return &chainStorage{
		persistent: persistent,
		memory:     memory,
	}
}

type chainStorage struct {
	persistent Storage
	memory     Storage
}

func (s *chainStorage) Write(collection, key string, v any) error {
	err := s.persistent.Write(collection, key, v)
	err2 := s.memory.Write(collection, key, v)
	if err != nil {
		return err
	}

	return err2
}

func (s *chainStorage) Read(collection, key string, v any) error {
	err := s.memory.Read(collection, key, v)
	if err == nil {
		return nil
	}

	err = s.persistent.Read(collection, key, v)
	if err != nil {
		// cache persistent data to memory to have faster access as well
		_ = s.memory.Write(collection, key, v)
	}

	return err
}

func (s *chainStorage) GetKeys(collection string) ([]string, error) {
	return s.persistent.GetKeys(collection)
}

func (s *chainStorage) Delete(collection, key string) error {
	err := s.persistent.Delete(collection, key)
	err2 := s.memory.Delete(collection, key)
	if err != nil {
		return err
	}

	return err2
}
