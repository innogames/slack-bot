package storage

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
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

var redisCtx = context.Background()

func (s redisStorage) Write(collection, key string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	s.client.HSet(redisCtx, collection, key, data)

	return nil
}

func (s redisStorage) Read(collection, key string, v any) error {
	res, err := s.client.HGet(redisCtx, collection, key).Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(res, v)
}

func (s redisStorage) GetKeys(collection string) ([]string, error) {
	res, err := s.client.HKeys(redisCtx, collection).Result()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s redisStorage) Delete(collection, key string) error {
	return s.client.HDel(redisCtx, collection, key).Err()
}
