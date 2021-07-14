package storage

import (
	"encoding/json"

	"github.com/go-redis/redis/v7"
)

// NewRedisStorage defined a redis bases storage to persist bot related information
func NewRedisStorage(client *redis.Client) Storage {
	return &redisStorage{
		client: client,
	}
}

type redisStorage struct {
	client *redis.Client
}

func (s redisStorage) Write(collection, key string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	s.client.HSet(collection, key, string(data))

	return nil
}

func (s redisStorage) Read(collection, key string, v interface{}) error {
	res, err := s.client.HGet(collection, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(res), v)
}

func (s redisStorage) GetKeys(collection string) ([]string, error) {
	res, err := s.client.HKeys(collection).Result()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s redisStorage) Delete(collection, key string) error {
	_, err := s.client.HDel(collection, key).Result()

	return err
}
