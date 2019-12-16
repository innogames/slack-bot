package storage

func newRedisStorage() (storage, error) {
	// todo implement
	return &redisStorage{}, nil
}

type redisStorage struct {
}

func (s redisStorage) Write(collection, resource string, v interface{}) error {
	panic("implement me")
}

func (s redisStorage) Read(collection, resource string, v interface{}) error {
	panic("implement me")
}

func (s redisStorage) GetKeys(collection string) ([]string, error) {
	panic("implement me")
}

func (s redisStorage) Delete(collection, resource string) error {
	panic("implement me")
}
